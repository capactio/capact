package controller

import (
	"context"
	"encoding/json"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"time"

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

// temporaryBuiltinArgoRunnerName represent the Argo Workflow runner interface which is temporary treated
// as built-in runner.
const temporaryBuiltinArgoRunnerName = "cap.interface.runner.argo"

// ActionService provides business functionality for reconciling Action CR.
type ActionService struct {
	k8sCli        client.Client
	runnerTimeout time.Duration
}

// NewActionService return new ActionService instance.
func NewActionService(cli client.Client, runnerTimeout time.Duration) *ActionService {
	return &ActionService{
		k8sCli:        cli,
		runnerTimeout: runnerTimeout,
	}
}

// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=create
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=create

// EnsureWorkflowSAExists creates dedicated ServiceAccount with cluster-admin permissions.
//
// This method MUST be removed in the near future as we should use a user service account instead.
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

	if err := a.k8sCli.Create(ctx, sa); a.ignoreAlreadyExits(err) != nil {
		return nil, errors.Wrap(err, "while creating service account")
	}

	if err := a.k8sCli.Create(ctx, binding); a.ignoreAlreadyExits(err) != nil {
		return nil, errors.Wrap(err, "while creating binding")
	}

	return sa, nil
}

// EnsureRunnerExecuted ensures that Kubernetes Job is created.
func (a *ActionService) EnsureRunnerExecuted(ctx context.Context, saName string, action *v1alpha1.Action) error {
	renderedAction, err := a.extractRunnerInterfaceAndArgs(action)
	if err != nil {
		return errors.Wrap(err, "while extracting rendered action from raw form")
	}

	switch renderedAction.RunnerInterface {
	case temporaryBuiltinArgoRunnerName:
		runnerJob := a.argoRunnerJob(saName, action)

		err = a.k8sCli.Create(ctx, runnerJob)
		return a.ignoreAlreadyExits(err)
	default:
		return errors.Errorf("unsupported %q runner", renderedAction.RunnerInterface)
	}
}

// EnsureRunnerInputDataCreated ensures that Kubernetes Secret with input data for a runner is created.
func (a *ActionService) EnsureRunnerInputDataCreated(ctx context.Context, saName string, action *v1alpha1.Action) error {
	renderedAction, err := a.extractRunnerInterfaceAndArgs(action)
	if err != nil {
		return errors.Wrap(err, "while extracting rendered action from raw form")
	}

	runnerInput := runner.InputData{
		Context: runner.ExecutionContext{
			Name:    action.Name,
			DryRun:  action.Spec.IsDryRun(),
			Timeout: 0,
			Platform: runner.KubernetesPlatformConfig{
				Namespace:          action.Namespace,
				ServiceAccountName: saName,
			},
		},
		Args: renderedAction.Args,
	}

	if action.Spec.DryRun != nil {
		runnerInput.Context.DryRun = *action.Spec.DryRun
	}

	marshaledInput, err := yaml.Marshal(runnerInput)
	if err != nil {
		return errors.Wrap(err, "while marshaling runner input data")
	}

	secret := &corev1.Secret{
		ObjectMeta: a.objectMetaFromAction(action),
		Data: map[string][]byte{
			"input.yaml": marshaledInput,
		},
	}

	err = a.k8sCli.Create(ctx, secret)
	switch {
	case err == nil:
	case apierrors.IsAlreadyExists(err):
		var oldSecret corev1.Secret
		key := client.ObjectKey{Name: secret.Name, Namespace: secret.Namespace}
		if err := a.k8sCli.Get(ctx, key, &oldSecret); err != nil {
			return err
		}

		if !metav1.IsControlledBy(&oldSecret, action) {
			return errors.Errorf("secret with the name %s already exists and it is not owned by Action with the same name", key.String())
		}
		// Do not mutate as the job can be already in running state?
		//	oldSecret.Data = secret.Data
		return nil
	default:
		return err
	}

	return nil
}

func (a *ActionService) EnsureRunnerStatusIsUpToDate(ctx context.Context, action *v1alpha1.Action) error {
	secret := &corev1.Secret{}
	key := client.ObjectKey{Name: action.Name, Namespace: action.Namespace}
	if err := a.k8sCli.Get(ctx, key, secret); err != nil {
		return errors.Wrap(err, "while getting secret with status")
	}

	if secret.Data == nil {
		return nil
	}

	status, found := secret.Data["status"]
	if !found {
		return nil
	}

	if action.Status.Runner == nil {
		action.Status.Runner = &v1alpha1.RunnerStatus{
			Interface: "why.we.need.that.?",
		}
	}

	if action.Status.Runner.Status != nil && reflect.DeepEqual(action.Status.Runner.Status.Raw, status) {
		return nil
	}

	action.Status.Runner.Status = &runtime.RawExtension{
		Raw: status,
	}
	if err := a.k8sCli.Status().Update(ctx, action); err != nil {
		return errors.Wrap(err, "while updating status of executed action")
	}

	return nil
}

type GetRunnerJobStatusOutput struct {
	Finished  bool
	JobStatus batchv1.JobConditionType
}

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

func (a *ActionService) ignoreAlreadyExits(err error) error {
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	return nil
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
	return &batchv1.Job{
		ObjectMeta: a.objectMetaFromAction(action),
		Spec: batchv1.JobSpec{
			BackoffLimit: ptr.Int32(0),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: saName,
					RestartPolicy:      corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "runner",
							Image: "gcr.io/projectvoltron/pr/argo-runner:engine",
							Env: []corev1.EnvVar{
								{
									Name:  "RUNNER_INPUT_PATH",
									Value: "/mnt/input.yaml",
								},
								{
									Name:  "RUNNER_LOGGER_DEV_MODE",
									Value: "true",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "input-volume",
									MountPath: "/mnt",
								},
							},
							ImagePullPolicy: "Never", // TODO: Always
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "input-volume",
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

type RenderedAction struct {
	RunnerInterface string          `json:"runnerInterface"`
	Args            json.RawMessage `json:"args"`
}

// extractRunnerInterfaceAndArgs
// assumption the `runnerInterface` is already resolved, currently we do not support revision
func (a *ActionService) extractRunnerInterfaceAndArgs(action *v1alpha1.Action) (*RenderedAction, error) {
	var renderingAction RenderedAction
	err := yaml.Unmarshal(action.Status.Rendering.Action.Raw, &renderingAction)
	if err != nil {
		return nil, err
	}

	return &renderingAction, nil
}
