package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	graphqldomain "capact.io/capact/internal/k8s-engine/graphql/domain/action"
	policypkg "capact.io/capact/internal/k8s-engine/policy"
	statusreporter "capact.io/capact/internal/k8s-engine/status-reporter"
	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/api/v1alpha1"
	"capact.io/capact/pkg/engine/k8s/policy"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	"capact.io/capact/pkg/hub/client/local"
	"capact.io/capact/pkg/runner"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/renderer/argo"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"
)

const (
	// temporaryBuiltinArgoRunnerName represent the Argo Workflow runner interface which is temporary treated
	// as built-in runner.
	temporaryBuiltinArgoRunnerName = "cap.interface.runner.argo.run"

	// #nosec G101
	runnerArgsSecretKey            = "args.yaml"
	runnerContextSecretKey         = "context.yaml"
	k8sJobRunnerInputDataMountPath = "/mnt"
	k8sJobRunnerVolumeName         = "input-volume"
	k8sJobActiveDeadlinePadding    = 10 * time.Second
)

type (
	// ArgoRenderer allows to render Capact Action defines in Argo format.
	ArgoRenderer interface {
		Render(ctx context.Context, input *argo.RenderInput) (*argo.RenderOutput, error)
	}
	// ActionValidator allows to validate Action definition.
	ActionValidator interface {
		Validate(action *types.Action, namespace string) error
	}
	// PolicyService allows to manage Capact Policy.
	PolicyService interface {
		Get(ctx context.Context) (policy.Policy, error)
	}
	// TypeInstanceLocker allows to lock and unlock given TypeInstances.
	TypeInstanceLocker interface {
		LockTypeInstances(ctx context.Context, in *gqllocalapi.LockTypeInstancesInput) error
		UnlockTypeInstances(ctx context.Context, in *gqllocalapi.UnlockTypeInstancesInput) error
	}
	// TypeInstanceGetter allow to fetch given TypeInstances from Hub.
	TypeInstanceGetter interface {
		ListTypeInstances(ctx context.Context, filter *gqllocalapi.TypeInstanceFilter, opts ...local.TypeInstancesOption) ([]gqllocalapi.TypeInstance, error)
	}
)

// ActionService provides business functionality for reconciling Action CR.
type ActionService struct {
	k8sCli             client.Client
	builtinRunner      BuiltinRunnerConfig
	argoRenderer       ArgoRenderer
	actionValidator    ActionValidator
	policyService      PolicyService
	policyOrder        policy.MergeOrder
	typeInstanceLocker TypeInstanceLocker
	typeInstanceGetter TypeInstanceGetter
	log                *zap.Logger
}

// NewActionService return new ActionService instance.
func NewActionService(log *zap.Logger, cli client.Client, argoRenderer ArgoRenderer, actionValidator ActionValidator, policyService PolicyService, policyOrder policy.MergeOrder, typeInstanceLocker TypeInstanceLocker, typeInstanceGetter TypeInstanceGetter, cfg Config) *ActionService {
	return &ActionService{
		k8sCli:             cli,
		builtinRunner:      cfg.BuiltinRunner,
		argoRenderer:       argoRenderer,
		actionValidator:    actionValidator,
		policyService:      policyService,
		policyOrder:        policyOrder,
		typeInstanceLocker: typeInstanceLocker,
		typeInstanceGetter: typeInstanceGetter,
		log:                log,
	}
}

// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=create
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=create

// EnsureWorkflowSAExists creates dedicated ServiceAccount with cluster-admin permissions.
//
// TODO: This method MUST be removed in the near future as we should use a user service account instead.
// When deleting, remove also the above kubebuilder rbac markers.
func (a *ActionService) EnsureWorkflowSAExists(ctx context.Context, action *v1alpha1.Action) (*corev1.ServiceAccount, error) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: a.objectMetaFromAction(action),
	}

	binding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: a.objectMetaFromAction(action),
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      sa.Name,
				Namespace: sa.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "cluster-admin",
		},
	}

	err := a.k8sCli.Create(ctx, sa)
	switch {
	case err == nil:
	case apierrors.IsAlreadyExists(err):
		old := &corev1.ServiceAccount{}
		key := client.ObjectKey{Name: sa.Name, Namespace: sa.Namespace}
		if err := a.k8sCli.Get(ctx, key, old); err != nil {
			return nil, err
		}

		if !metav1.IsControlledBy(old, action) {
			return nil, errors.Errorf("ServiceAccount %q already exists and it is not owned by Action with the same name", key.String())
		}
	default:
		return nil, errors.Wrap(err, "while creating service account")
	}

	err = a.k8sCli.Create(ctx, binding)
	switch {
	case err == nil:
	case apierrors.IsAlreadyExists(err):
		old := &rbacv1.ClusterRoleBinding{}
		key := client.ObjectKey{Name: binding.Name}
		if err := a.k8sCli.Get(ctx, key, old); err != nil {
			return nil, err
		}

		if !metav1.IsControlledBy(old, action) {
			return nil, errors.Errorf("RoleBinding %q already exists and it is not owned by Action with the same name", key.String())
		}
		old.Subjects = binding.Subjects
		old.RoleRef = binding.RoleRef
		if err := a.k8sCli.Update(ctx, old); err != nil {
			return nil, err
		}
	default:
		return nil, errors.Wrap(err, "while creating binding")
	}

	return sa, nil
}

// CleanupActionOwnedResources removes the Action owned resources.
func (a *ActionService) CleanupActionOwnedResources(ctx context.Context, action *v1alpha1.Action) (bool, error) {
	isCleanupIgnored := true
	if !controllerutil.ContainsFinalizer(action, v1alpha1.ActionFinalizer) {
		return isCleanupIgnored, nil // our finalizer was already removed
	}

	if action.IsExecuted() {
		// Current decision:
		a.log.Info("Ignoring delete request. Wait until Action execution will be finished.", zap.String("phase", string(action.Status.Phase)))

		// Deletion in this state is complicated as such Action can be in the middle of e.g. data migration, system shutdown,
		// creating resources on hyperscaler side, etc.
		// In the future we can revisit this approach based on user feedback and e.g. cancel running actions, maybe even rollback
		// already executed steps, etc.
		return isCleanupIgnored, nil
	}

	// ===== Execute clean-up logic
	isCleanupIgnored = false

	// 1. Ensure that TypeInstances are unlocked.
	err := a.UnlockTypeInstances(ctx, action)
	if a.ignoreNotActionableTypeInstanceErrors(err) != nil {
		return isCleanupIgnored, errors.Wrap(err, "while unlocking TypeInstances")
	}

	// 2. Ensure ClusterRoleBinding deleted
	// We use ownerReference for created resources. But the cluster-scoped resources cannot have namespace-scoped owners.
	// As a result, we need to remove it manually.
	binding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: a.objectMetaFromAction(action),
	}
	if err := a.k8sCli.Delete(ctx, binding); client.IgnoreNotFound(err) != nil {
		return isCleanupIgnored, errors.Wrapf(err, "while deleting ClusterRoleBinding owned by %s/%s Action", action.GetName(), action.GetNamespace())
	}

	// 3. Remove finalizer
	controllerutil.RemoveFinalizer(action, v1alpha1.ActionFinalizer)
	if err := a.k8sCli.Update(ctx, action); err != nil {
		return isCleanupIgnored, errors.Wrap(err, "while removing Action finalizer")
	}

	return isCleanupIgnored, nil
}

// ignoreNotActionableTypeInstanceErrors ignores GraphQL error which says that TI are locked by different owner or do not exist.
// In our case it means that TI were already unlocked by a given Action and someone else locked them or deleted.
//
// TODO: Get rid of ridiculous string assertion after adding proper error types to GraphQL responses.
//       http://knowyourmeme.com/memes/this-is-fine
func (a *ActionService) ignoreNotActionableTypeInstanceErrors(err error) error {
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "locked by different owner") {
		return nil
	}

	if strings.Contains(err.Error(), "not found") {
		return nil
	}

	return err
}

// EnsureRunnerInputDataCreated ensures that Kubernetes Secret with input data for a runner is created and up to date.
func (a *ActionService) EnsureRunnerInputDataCreated(ctx context.Context, saName string, action *v1alpha1.Action) error {
	runnerCtx := runner.Context{
		Name:    action.Name,
		DryRun:  action.Spec.IsDryRun(),
		Timeout: runner.Duration(a.builtinRunner.Timeout),
		Platform: runner.KubernetesPlatformConfig{
			Namespace:          action.Namespace,
			ServiceAccountName: saName,
			OwnerRef:           *metav1.NewControllerRef(action, v1alpha1.GroupVersion.WithKind(v1alpha1.ActionKind)),
		},
	}

	marshalledRunnerCtx, err := yaml.Marshal(runnerCtx)
	if err != nil {
		return errors.Wrap(err, "while marshaling runner context")
	}

	renderedAction, err := a.extractRunnerInterfaceAndArgs(action)
	if err != nil {
		return errors.Wrap(err, "while extracting rendered action from raw form")
	}
	marshalledRunnerArgs, err := yaml.Marshal(renderedAction.Args)
	if err != nil {
		return errors.Wrap(err, "while marshaling runner args")
	}

	secret := &corev1.Secret{
		ObjectMeta: a.objectMetaFromAction(action),
		Data: map[string][]byte{
			runnerContextSecretKey: marshalledRunnerCtx,
			runnerArgsSecretKey:    marshalledRunnerArgs,
		},
	}

	err = a.k8sCli.Create(ctx, secret)
	switch {
	case err == nil:
	case apierrors.IsAlreadyExists(err):
		oldSecret := &corev1.Secret{}
		key := client.ObjectKey{Name: secret.Name, Namespace: secret.Namespace}
		if err := a.k8sCli.Get(ctx, key, oldSecret); err != nil {
			return err
		}

		oldSecret.Data[runnerContextSecretKey] = secret.Data[runnerContextSecretKey]
		oldSecret.Data[runnerArgsSecretKey] = secret.Data[runnerArgsSecretKey]
		return a.k8sCli.Update(ctx, oldSecret)
	default:
		return err
	}

	return nil
}

// EnsureRunnerExecuted ensures that Kubernetes Job is created and up to date.
func (a *ActionService) EnsureRunnerExecuted(ctx context.Context, saName string, action *v1alpha1.Action) error {
	renderedAction, err := a.extractRunnerInterfaceAndArgs(action)
	if err != nil {
		return errors.Wrap(err, "while extracting rendered action from raw form")
	}

	// TODO: Change that to generic option similar to k8s plugins which can be registered from separate pkg
	// example: https://github.com/kubernetes/kubernetes/blob/v1.19.4/pkg/kubeapiserver/options/plugins.go
	if renderedAction.RunnerInterface != temporaryBuiltinArgoRunnerName {
		return errors.Errorf("unsupported %q runner", renderedAction.RunnerInterface)
	}

	runnerJob := a.argoRunnerJob(saName, action)

	err = a.k8sCli.Create(ctx, runnerJob)
	switch {
	case err == nil:
	case apierrors.IsAlreadyExists(err):
		old := &batchv1.Job{}
		key := client.ObjectKey{Name: runnerJob.Name, Namespace: runnerJob.Namespace}
		if err := a.k8sCli.Get(ctx, key, old); err != nil {
			return err
		}

		if !metav1.IsControlledBy(old, action) {
			return errors.Errorf("secret with the name %s already exists and it is not owned by Action with the same name", key.String())
		}
	default:
		return err
	}

	return nil
}

// LockTypeInstances locks TypeInstance used by a given Action.
func (a *ActionService) LockTypeInstances(ctx context.Context, action *v1alpha1.Action) error {
	if action == nil || action.Status.Rendering == nil {
		return errors.New("Action or Action rendering status is nil")
	}

	if action.Status.Rendering.TypeInstancesToLock == nil {
		return nil
	}

	ownerID := ownerIDKey(action)

	return a.typeInstanceLocker.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
		OwnerID: ownerID,
		Ids:     action.Status.Rendering.TypeInstancesToLock,
	})
}

// UnlockTypeInstances unlocks TypeInstances used by a given Action.
func (a *ActionService) UnlockTypeInstances(ctx context.Context, action *v1alpha1.Action) error {
	if action == nil || action.Status.Rendering == nil || action.Status.Rendering.TypeInstancesToLock == nil {
		return nil
	}

	ownerID := ownerIDKey(action)

	return a.typeInstanceLocker.UnlockTypeInstances(ctx, &gqllocalapi.UnlockTypeInstancesInput{
		OwnerID: ownerID,
		Ids:     action.Status.Rendering.TypeInstancesToLock,
	})
}

// RenderAction returns rendered Implementation for Interface from a given Action.
func (a *ActionService) RenderAction(ctx context.Context, action *v1alpha1.Action) (*v1alpha1.RenderingStatus, error) {
	ref, parametersCollection, err := a.getUserInputData(ctx, action)
	if err != nil {
		return nil, err
	}

	actionPolicy, actionPolicyData, err := a.getActionPolicyData(ctx, action)
	if err != nil {
		return nil, err
	}

	typeInstances := a.getUserInputTypeInstances(action)

	runnerCtxSecretRef := argo.RunnerContextSecretRef{
		Name: action.Name,
		Key:  runnerContextSecretKey,
	}
	interfaceRef := types.InterfaceRef{
		Path:     string(action.Spec.ActionRef.Path),
		Revision: action.Spec.ActionRef.Revision,
	}

	policy, err := a.getPolicyWithFallbackToEmpty(ctx)
	if err != nil {
		return nil, err
	}

	ownerID := ownerIDKey(action)
	options := []argo.RendererOption{
		argo.WithSecretUserInput(ref, parametersCollection),
		argo.WithPolicyOrder(a.policyOrder),
		argo.WithGlobalPolicy(policy),
		argo.WithTypeInstances(typeInstances),
		argo.WithOwnerID(ownerID),
	}

	if actionPolicy != nil {
		options = append(options, argo.WithActionPolicy(*actionPolicy))
	}

	renderOutput, err := a.argoRenderer.Render(
		ctx,
		&argo.RenderInput{
			RunnerContextSecretRef: runnerCtxSecretRef,
			InterfaceRef:           interfaceRef,
			Options:                options,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "while rendering Action")
	}

	actionBytes, err := json.Marshal(renderOutput.Action)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling action to json")
	}

	parametersBytes, err := json.Marshal(parametersCollection)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling user input to json")
	}

	status := &v1alpha1.RenderingStatus{}
	status.SetAction(actionBytes)
	status.SetInputParameters(parametersBytes)
	status.SetTypeInstancesToLock(renderOutput.TypeInstancesToLock)
	status.SetActionPolicy(actionPolicyData)

	if err := a.actionValidator.Validate(renderOutput.Action, action.Namespace); err != nil {
		return status, errors.Wrap(err, "while validating rendered Action")
	}

	return status, nil
}

func (a *ActionService) getUserInputData(ctx context.Context, action *v1alpha1.Action) (*argo.UserInputSecretRef, types.ParametersCollection, error) {
	if action.Spec.Input == nil || action.Spec.Input.Parameters == nil {
		return nil, nil, nil
	}

	secret := &corev1.Secret{}
	key := client.ObjectKey{Name: action.Spec.Input.Parameters.SecretRef.Name, Namespace: action.Namespace}
	if err := a.k8sCli.Get(ctx, key, secret); err != nil {
		return nil, nil, errors.Wrap(err, "while getting K8s Secret with user input data")
	}

	parameters := types.ParametersCollection{}
	for key, data := range secret.Data {
		if ok, name := graphqldomain.IsParameterDataKey(key); ok {
			parameters[name] = string(data)
		}
	}

	return &argo.UserInputSecretRef{
		Name: action.Spec.Input.Parameters.SecretRef.Name,
	}, parameters, nil
}

func (a *ActionService) getActionPolicyData(ctx context.Context, action *v1alpha1.Action) (*policy.ActionPolicy, []byte, error) {
	if action.Spec.Input == nil || action.Spec.Input.ActionPolicy == nil {
		return nil, nil, nil
	}

	secret := &corev1.Secret{}
	key := client.ObjectKey{Name: action.Spec.Input.ActionPolicy.SecretRef.Name, Namespace: action.Namespace}
	if err := a.k8sCli.Get(ctx, key, secret); err != nil {
		return nil, nil, errors.Wrap(err, "while getting K8s Secret with user input data")
	}

	policyData := secret.Data[graphqldomain.ActionPolicySecretDataKey]

	policy := &policy.ActionPolicy{}
	if err := json.Unmarshal(policyData, policy); err != nil {
		return nil, nil, errors.Wrap(err, "while unmarshaling Policy data")
	}

	return policy, policyData, nil
}

func (a *ActionService) getUserInputTypeInstances(action *v1alpha1.Action) []types.InputTypeInstanceRef {
	if action.Spec.Input == nil || action.Spec.Input.TypeInstances == nil {
		return nil
	}

	var refs []types.InputTypeInstanceRef
	for _, ti := range *action.Spec.Input.TypeInstances {
		refs = append(refs, types.InputTypeInstanceRef{Name: ti.Name, ID: ti.ID})
	}

	return refs
}

func (a *ActionService) getPolicyWithFallbackToEmpty(ctx context.Context) (policy.Policy, error) {
	p, err := a.policyService.Get(ctx)
	if err != nil {
		if errors.Is(err, policypkg.ErrPolicyConfigMapNotFound) {
			a.log.Info("ConfigMap with cluster policy not found. Fallback to empty Cluster Policy")
			return policy.Policy{}, nil
		}

		return policy.Policy{}, errors.Wrap(err, "while getting K8s ConfigMap with cluster policy")
	}

	return p, nil
}

// GetReportedRunnerStatusOutput defines output for GetReportedRunnerStatus method.
type GetReportedRunnerStatusOutput struct {
	Changed bool
	Status  []byte
}

// GetReportedRunnerStatus returns status reported by action runner.
func (a *ActionService) GetReportedRunnerStatus(ctx context.Context, action *v1alpha1.Action) (*GetReportedRunnerStatusOutput, error) {
	// TODO: consider to move logic with fetching current status to status-reporter pkg
	secret := &corev1.Secret{}
	key := client.ObjectKey{Name: action.Name, Namespace: action.Namespace}
	if err := a.k8sCli.Get(ctx, key, secret); err != nil {
		return nil, errors.Wrap(err, "while getting secret with status")
	}

	if secret.Data == nil {
		return &GetReportedRunnerStatusOutput{Changed: false}, nil
	}

	status, found := secret.Data[statusreporter.SecretStatusEntryKey]
	if !found {
		return &GetReportedRunnerStatusOutput{Changed: false}, nil
	}

	if action.Status.Runner != nil && action.Status.Runner.Status != nil &&
		bytes.Equal(action.Status.Runner.Status.Raw, status) {
		return &GetReportedRunnerStatusOutput{Changed: false}, nil
	}

	return &GetReportedRunnerStatusOutput{
		Changed: true,
		Status:  status,
	}, nil
}

// GetRunnerJobStatusOutput defines output for GetRunnerJobStatus method.
type GetRunnerJobStatusOutput struct {
	Finished  bool
	JobStatus batchv1.JobConditionType
}

// GetRunnerJobStatus returns K8s Job status which executes action runner.
func (a *ActionService) GetRunnerJobStatus(ctx context.Context, action *v1alpha1.Action) (*GetRunnerJobStatusOutput, error) {
	runnerJob := &batchv1.Job{}
	key := client.ObjectKey{Name: action.Name, Namespace: action.Namespace}
	if err := a.k8sCli.Get(ctx, key, runnerJob); err != nil {
		return nil, errors.Wrap(err, "while getting runner k8s job")
	}

	status, finished := jobFinishStatus(runnerJob)
	return &GetRunnerJobStatusOutput{
		Finished:  finished,
		JobStatus: status,
	}, nil
}

func jobFinishStatus(j *batchv1.Job) (batchv1.JobConditionType, bool) {
	for _, c := range j.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			return c.Type, true
		}
	}
	return "", false
}

// GetTypeInstancesFromAction returns TypeInstances created by a given Action.
func (a *ActionService) GetTypeInstancesFromAction(ctx context.Context, action *v1alpha1.Action) ([]v1alpha1.OutputTypeInstanceDetails, error) {
	ownerID := ownerIDKey(action)

	typeInstances, err := a.typeInstanceGetter.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{
		CreatedBy: &ownerID,
	}, local.WithFields(local.TypeInstanceRootFields|local.TypeInstanceTypeRefFields))
	if err != nil {
		return nil, errors.Wrap(err, "while listing TypeInstances")
	}

	var res []v1alpha1.OutputTypeInstanceDetails
	for _, ti := range typeInstances {
		res = append(res, v1alpha1.OutputTypeInstanceDetails{
			CommonTypeInstanceDetails: v1alpha1.CommonTypeInstanceDetails{
				ID: ti.ID,
				TypeRef: &v1alpha1.ManifestReference{
					Path:     v1alpha1.NodePath(ti.TypeRef.Path),
					Revision: &ti.TypeRef.Revision,
				},
			},
		})
	}

	return res, nil
}

// objectMetaFromAction uses given Action Name and Namespace, to set the same values on new ObjectMeta.
// Additionally, sets ownerReference to a given Action.
//
// In the future we can set `GenerateName = action.Name`, to remove problem with name collisions.
// With such change, we will need to introduce an informer indexer to be able to get objects with a generated names.
// Example indexer:
// https://github.com/kubernetes-sigs/kubebuilder/blob/8823b61390eca446c9f44542f5a44309941a62a3/docs/book/src/cronjob-tutorial/testdata/project/controllers/cronjob_controller.go#L548
func (a *ActionService) objectMetaFromAction(action *v1alpha1.Action) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      action.Name,
		Namespace: action.Namespace,
		OwnerReferences: []metav1.OwnerReference{
			*metav1.NewControllerRef(action, v1alpha1.GroupVersion.WithKind(v1alpha1.ActionKind)),
		},
	}
}

func (a *ActionService) argoRunnerJob(saName string, action *v1alpha1.Action) *batchv1.Job {
	activeDeadline := a.builtinRunner.Timeout + k8sJobActiveDeadlinePadding
	activeDeadlineSec := activeDeadline.Seconds()

	return &batchv1.Job{
		ObjectMeta: a.objectMetaFromAction(action),
		Spec: batchv1.JobSpec{
			// TODO: In the future we should add retries and each runner should handle them properly.
			BackoffLimit: ptr.Int32(0),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName:    saName,
					RestartPolicy:         corev1.RestartPolicyNever,
					ActiveDeadlineSeconds: ptr.Int64(int64(activeDeadlineSec)),
					Containers: []corev1.Container{
						{
							Name:  "runner",
							Image: a.builtinRunner.Image,
							Env: []corev1.EnvVar{
								{
									Name:  "RUNNER_ARGS_PATH",
									Value: fmt.Sprintf("%s/%s", k8sJobRunnerInputDataMountPath, runnerArgsSecretKey),
								},
								{
									Name:  "RUNNER_CONTEXT_PATH",
									Value: fmt.Sprintf("%s/%s", k8sJobRunnerInputDataMountPath, runnerContextSecretKey),
								},
								{
									Name:  "RUNNER_LOGGER_DEV_MODE",
									Value: "true",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      k8sJobRunnerVolumeName,
									MountPath: k8sJobRunnerInputDataMountPath,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: k8sJobRunnerVolumeName,
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: action.Name,
									Optional:   ptr.Bool(false),
								},
							},
						},
					},
				},
			},
		},
	}
}

type renderedAction struct {
	RunnerInterface string          `json:"runnerInterface"`
	Args            json.RawMessage `json:"args"`
}

// CAUTION: assumption that the `runnerInterface` is already resolved to full node path. Currently, we do not support revision.
func (a *ActionService) extractRunnerInterfaceAndArgs(action *v1alpha1.Action) (*renderedAction, error) {
	var renderingAction renderedAction
	err := yaml.Unmarshal(action.Status.Rendering.Action.Raw, &renderingAction)
	if err != nil {
		return nil, err
	}

	return &renderingAction, nil
}

func ownerIDKey(a *v1alpha1.Action) string {
	return fmt.Sprintf("%s/%s-%s", a.Namespace, a.Name, a.UID)
}
