package controller

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	authv1 "k8s.io/api/authentication/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"projectvoltron.dev/voltron/internal/ptr"
	corev1alpha1 "projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
	ochgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// ActionReconciler reconciles a Action object.
type ActionReconciler struct {
	client.Client
	Log logr.Logger
	gatewayInterface
}

type gatewayInterface interface {
	GetImplementationLatestRevision(ctx context.Context, path string) (*ochgraphql.ImplementationRevision, error)
}

// NewActionReconciler returns the ActionReconciler instance.
func NewActionReconciler(client client.Client, log logr.Logger, gatewayInterface gatewayInterface) *ActionReconciler {
	return &ActionReconciler{Client: client, Log: log, gatewayInterface: gatewayInterface}
}

// +kubebuilder:rbac:groups=core.projectvoltron.dev,resources=actions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.projectvoltron.dev,resources=actions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=create
// +kubebuilder:rbac:groups="",resources=secrets,verbs=create

// Reconcile handles the reconcile logic for the Action CR.
func (r *ActionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var (
		ctx = context.Background()
		log = r.Log.WithValues("action", req.NamespacedName)
	)

	// Just a simple logic to check if controller is working e2e
	// TODO: replace in https://cshark.atlassian.net/browse/SV-34

	var action corev1alpha1.Action
	if err := r.Get(ctx, req.NamespacedName, &action); err != nil {
		log.Error(err, "while fetching Action CR")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("action", "action", action)

	if action.Status.Phase == corev1alpha1.CreatedActionPhase || action.Status.Phase == corev1alpha1.InitialActionPhase {
		err := r.renderAction(ctx, log.WithValues("phase", "renderAction"), &action)
		return ctrl.Result{}, err
	}

	if action.Status.Phase == corev1alpha1.ReadyToRunActionPhase {
		job := r.dummyJob(action)
		if err := r.Create(ctx, &job); err != nil {
			log.Error(err, "while creating dummy Job")
			return ctrl.Result{}, err
		}

		r.setSampleStatus(&action)
		err := r.Status().Update(ctx, &action)
		if err != nil {
			log.Error(err, "while updating Action CR status")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *ActionReconciler) renderAction(ctx context.Context, log logr.Logger, action *corev1alpha1.Action) error {
	log.Info("rendering workflow")

	latestRevision, err := r.gatewayInterface.GetImplementationLatestRevision(ctx, string(action.Spec.Path))
	if err != nil {
		return errors.Wrap(err, "cannot fetch implementation via gateway")
	}

	if latestRevision == nil || latestRevision.Spec == nil ||
		latestRevision.Spec.Action == nil {
		return errors.Wrap(err, "missing action workflow in fetched implementation")
	}

	actionBytes, err := json.Marshal(latestRevision.Spec.Action)
	if err != nil {
		return errors.Wrap(err, "failed to marshal action to json")
	}

	if action.Status.Rendering == nil {
		action.Status.Rendering = &corev1alpha1.RenderingStatus{}
	}

	action.Status.Rendering.Action = &runtime.RawExtension{
		Raw: actionBytes,
	}
	action.Status.Phase = corev1alpha1.BeingRenderedActionPhase
	action.Status.LastTransitionTime = metav1.Now()

	if err := r.Status().Update(ctx, action); err != nil {
		return errors.Wrap(err, "failed to save updated action in k8s")
	}

	return nil
}

func (r *ActionReconciler) setSampleStatus(action *corev1alpha1.Action) {
	action.Status = corev1alpha1.ActionStatus{
		Phase:   corev1alpha1.SucceededActionPhase,
		Message: ptr.String("Foo"),
		Runner: &corev1alpha1.RunnerStatus{
			Interface: "cap.interface.runner.argo.run",
			Status: &runtime.RawExtension{
				Raw: []byte(`{"argoWorkflowRef": "sample"}`),
			},
		},
		Output: &corev1alpha1.ActionOutput{
			Artifacts: &[]corev1alpha1.OutputArtifactDetails{
				{
					CommonArtifactDetails: corev1alpha1.CommonArtifactDetails{
						Name:           "bar",
						TypeInstanceID: "b02bdc8e-9e5d-4ee0-a350-4ccc23b363fb",
						TypePath:       "cap.type.database.postgresql.config",
					},
				},
			},
		},
		Rendering: &corev1alpha1.RenderingStatus{
			Action: &runtime.RawExtension{
				Raw: []byte(`{"workflow": true}`),
			},
			Input: &corev1alpha1.ResolvedActionInput{
				Artifacts: &[]corev1alpha1.InputArtifactDetails{
					{
						CommonArtifactDetails: corev1alpha1.CommonArtifactDetails{
							Name:           "foo",
							TypeInstanceID: "fee33a5e-d957-488a-86bd-5dacd4120312",
							TypePath:       "cap.core.type.foo.bar",
						},
						Optional: false,
					},
					{
						CommonArtifactDetails: corev1alpha1.CommonArtifactDetails{
							Name:           "bar",
							TypeInstanceID: "563a79eb-7417-4e11-aa4b-d93076c04e48",
							TypePath:       "cap.core.type.bar.baz",
						},
						Optional: true,
					},
				},
				Parameters: &runtime.RawExtension{
					Raw: []byte(`{"input1": "foo", "input2": 2, "input3": { "nested": true }}`),
				},
			},
		},
		CreatedBy: &authv1.UserInfo{
			Username: "foo",
			UID:      "73d3c628-864e-45e3-8927-b9b71e17c110",
			Groups:   []string{"bar", "baz"},
		},
		RunBy: &authv1.UserInfo{
			Username: "bar",
			UID:      "3935025e-1403-4bb5-99d8-3ce428acf527",
			Groups:   []string{"bar", "baz"},
		},
		CancelledBy: &authv1.UserInfo{
			Username: "bar",
			UID:      "14354227-9afe-45c8-8808-765b6a7fcb2b",
			Groups:   []string{"bar", "baz"},
		},
		LastTransitionTime: metav1.Now(),
	}
}

func (r *ActionReconciler) dummyJob(action corev1alpha1.Action) batchv1.Job {
	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      action.Name,
			Namespace: action.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:    "runner",
							Image:   "alpine:latest",
							Command: []string{"echo", "hakuna-matata"},
						},
					},
				},
			},
		},
	}
}

func (r *ActionReconciler) SetupWithManager(mgr ctrl.Manager, maxConcurrentReconciles int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Action{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: maxConcurrentReconciles,
		}).
		Complete(r)
}
