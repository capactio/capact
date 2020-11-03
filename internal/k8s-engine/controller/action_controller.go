package controller

import (
	"context"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	corev1alpha1 "projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
)

// ActionReconciler reconciles a Action object.
type ActionReconciler struct {
	client.Client
	Log logr.Logger
}

// NewActionReconciler returns the ActionReconciler instance.
func NewActionReconciler(client client.Client, log logr.Logger) *ActionReconciler {
	return &ActionReconciler{Client: client, Log: log}
}

// +kubebuilder:rbac:groups=core.projectvoltron.dev,resources=actions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.projectvoltron.dev,resources=actions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=create

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

	job := r.dummyJob(action)
	if err := r.Create(ctx, &job); err != nil {
		log.Error(err, "while creating dummy Job")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
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
