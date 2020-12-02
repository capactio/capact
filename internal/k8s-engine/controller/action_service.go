package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	ochgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/public"

	statusreporter "projectvoltron.dev/voltron/internal/k8s-engine/status-reporter"
	"projectvoltron.dev/voltron/internal/ptr"
	"projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
	"projectvoltron.dev/voltron/pkg/runner"

	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	// temporaryBuiltinArgoRunnerName represent the Argo Workflow runner interface which is temporary treated
	// as built-in runner.
	temporaryBuiltinArgoRunnerName = "cap.interface.runner.argo"
	secretInputDataEntryName       = "input.yaml"
	k8sJobRunnerInputDataMountPath = "/mnt"
	k8sJobRunnerVolumeName         = "input-volume"
	k8sJobActiveDeadlinePadding    = 10 * time.Second
)

type OCHImplementationGetter interface {
	GetLatestRevisionOfImplementationForInterface(ctx context.Context, path string) (*ochgraphql.ImplementationRevision, error)
}

// ActionService provides business functionality for reconciling Action CR.
type ActionService struct {
	k8sCli             client.Client
	runnerTimeout      time.Duration
	builtinRunnerImage string
	implGetter         OCHImplementationGetter
}

// NewActionService return new ActionService instance.
func NewActionService(cli client.Client, implGetter OCHImplementationGetter, builtinRunnerImage string, runnerTimeout time.Duration) *ActionService {
	return &ActionService{
		k8sCli:             cli,
		runnerTimeout:      runnerTimeout,
		builtinRunnerImage: builtinRunnerImage,
		implGetter:         implGetter,
	}
}

// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=create
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=create

// EnsureWorkflowSAExists creates dedicated ServiceAccount with cluster-admin permissions.
//
// TODO: This method MUST be removed in the near future as we should use a user service account instead.
// When deleting, remove also the above kubebuilder rbac markers.
func (a *ActionService) EnsureWorkflowSAExists(ctx context.Context, action *v1alpha1.Action) (*corev1.ServiceAccount, error) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: a.objectMetaFromAction(action),
	}

	binding := &rbacv1.RoleBinding{
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
		old := &rbacv1.RoleBinding{}
		key := client.ObjectKey{Name: binding.Name, Namespace: binding.Namespace}
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

// EnsureRunnerInputDataCreated ensures that Kubernetes Secret with input data for a runner is created and up to date.
func (a *ActionService) EnsureRunnerInputDataCreated(ctx context.Context, saName string, action *v1alpha1.Action) error {
	renderedAction, err := a.extractRunnerInterfaceAndArgs(action)
	if err != nil {
		return errors.Wrap(err, "while extracting rendered action from raw form")
	}

	runnerInput := runner.InputData{
		Context: runner.ExecutionContext{
			Name:    action.Name,
			DryRun:  action.Spec.IsDryRun(),
			Timeout: runner.Duration(a.runnerTimeout),
			Platform: runner.KubernetesPlatformConfig{
				Namespace:          action.Namespace,
				ServiceAccountName: saName,
				OwnerRef:           *metav1.NewControllerRef(action, v1alpha1.GroupVersion.WithKind(v1alpha1.ActionKind)),
			},
		},
		Args: renderedAction.Args,
	}

	marshaledInput, err := yaml.Marshal(runnerInput)
	if err != nil {
		return errors.Wrap(err, "while marshaling runner input data")
	}

	secret := &corev1.Secret{
		ObjectMeta: a.objectMetaFromAction(action),
		Data: map[string][]byte{
			secretInputDataEntryName: marshaledInput,
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

		if !metav1.IsControlledBy(oldSecret, action) {
			return errors.Errorf("secret with the name %s already exists and it is not owned by Action with the same name", key.String())
		}
		oldSecret.Data = secret.Data
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

func (a *ActionService) isGCPSecretAvailable(ctx context.Context, action *v1alpha1.Action) bool {
	secret := &corev1.Secret{}
	key := client.ObjectKey{Name: "gcp-credentials", Namespace: action.Namespace}
	err := a.k8sCli.Get(ctx, key, secret)
	return err == nil
}

// ensureLocalSuffix adds the `-local` prefix if not already added
func (a *ActionService) ensureLocalSuffix(path string) string {
	name := filepath.Ext(path)
	prefix := strings.TrimSuffix(path, name)
	if !strings.HasSuffix(name, "-local") {
		name = name + "-local"
	}
	return prefix + name
}

// ResolveImplementationForAction returns specific implementation for interface from a given Action.
// TODO: This is a dummy implementation just for demo purpose.
func (a *ActionService) ResolveImplementationForAction(ctx context.Context, action *v1alpha1.Action) ([]byte, error) {
	path := string(action.Spec.ActionRef.Path)
	if !a.isGCPSecretAvailable(ctx, action) {
		path = a.ensureLocalSuffix(path)
	}

	latestRevision, err := a.implGetter.GetLatestRevisionOfImplementationForInterface(ctx, path)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching implementation")
	}

	if latestRevision == nil || latestRevision.Spec == nil || latestRevision.Spec.Action == nil {
		return nil, errors.New("missing action in Implementation revision")
	}

	actionBytes, err := json.Marshal(latestRevision.Spec.Action)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling action to json")
	}
	return actionBytes, nil
}

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
	activeDeadline := a.runnerTimeout + k8sJobActiveDeadlinePadding
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
							Image: a.builtinRunnerImage,
							Env: []corev1.EnvVar{
								{
									Name:  "RUNNER_INPUT_PATH",
									Value: fmt.Sprintf("%s/%s", k8sJobRunnerInputDataMountPath, secretInputDataEntryName),
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
