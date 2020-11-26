package controller

import (
	"context"
	"encoding/json"

	"projectvoltron.dev/voltron/internal/ptr"
	"projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
	corev1alpha1 "projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
	ochgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/public"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// ActionReconciler reconciles a Action object.
type ActionReconciler struct {
	client.Client
	log        logr.Logger
	svc        *ActionService // TODO interface
	recorder   record.EventRecorder
	implGetter OCHImplementationGetter
}

type OCHImplementationGetter interface {
	GetImplementationLatestRevision(ctx context.Context, path string) (*ochgraphql.ImplementationRevision, error)
}

// NewActionReconciler returns the ActionReconciler instance.
func NewActionReconciler(log logr.Logger, client client.Client, recorder record.EventRecorder, implementationGetter OCHImplementationGetter, svc *ActionService) *ActionReconciler {
	return &ActionReconciler{
		Client:     client,
		log:        log.WithName("controllers").WithName("Action"),
		svc:        svc,
		recorder:   recorder,
		implGetter: implementationGetter,
	}
}

// +kubebuilder:rbac:groups=core.projectvoltron.dev,resources=actions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.projectvoltron.dev,resources=actions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete

// Reconcile handles the reconcile logic for the Action CR.
// TODO: introduce and ignore permanent error in reconcile loop
func (r *ActionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var (
		ctx = context.Background()
		log = r.log.WithValues("action", req.NamespacedName)
	)

	// Just a simple logic to check if controller is working e2e
	// TODO: replace in https://cshark.atlassian.net/browse/SV-34

	action := &v1alpha1.Action{}
	if err := r.Get(ctx, req.NamespacedName, action); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "while fetching Action CR")
		return ctrl.Result{}, err
	}

	// TODO: add finalizer and handle deletion properly (cancel running actions, remove ArgoWorkflows)

	// TODO bug, that newly created Action CR has empty phase and not the default, so we need to handle it here
	if action.Status.Phase == "" || action.Status.Phase == corev1alpha1.InitialActionPhase {
		err := r.renderAction(ctx, log.WithValues("phase", "renderAction"), action)
		return ctrl.Result{}, err
	}

	if action.ShouldBeExecuted() {
		log.Info("Execute runner")
		if err := r.executeAction(ctx, action); err != nil {
			r.recorder.Event(action, corev1.EventTypeWarning, "Execute runner", err.Error())
			log.Error(err, "while executing runner")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if action.IsExecuted() {
		log.Info("Check runner status")
		if err := r.handleRunningAction(ctx, action); err != nil {
			r.recorder.Event(action, corev1.EventTypeWarning, "Check runner status", err.Error())
			log.Error(err, "while checking runner status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *ActionReconciler) renderAction(ctx context.Context, log logr.Logger, action *corev1alpha1.Action) error {
	log.Info("rendering workflow")

	action.Status.Phase = corev1alpha1.BeingRenderedActionPhase
	action.Status.LastTransitionTime = metav1.Now()
	action.Status.ObservedGeneration = action.Generation

	if err := r.Status().Update(ctx, action); err != nil {
		return errors.Wrap(err, "while updating action object status")
	}

	latestRevision, err := r.implGetter.GetImplementationLatestRevision(ctx, string(action.Spec.Path))
	if err != nil {
		return errors.Wrap(err, "while fetching implementation")
	}

	if latestRevision == nil || latestRevision.Spec == nil ||
		latestRevision.Spec.Action == nil {
		return errors.New("missing action in Implementation revision")
	}

	actionBytes, err := json.Marshal(latestRevision.Spec.Action)
	if err != nil {
		return errors.Wrap(err, "while marshaling action to json")
	}

	if action.Status.Rendering == nil {
		action.Status.Rendering = &corev1alpha1.RenderingStatus{}
	}

	action.Status.Rendering.Action = &runtime.RawExtension{
		Raw: actionBytes,
	}
	action.Status.Phase = corev1alpha1.ReadyToRunActionPhase
	action.Status.LastTransitionTime = metav1.Now()
	action.Status.ObservedGeneration = action.Generation

	log.Info("workflow rendered")
	if err := r.Status().Update(ctx, action); err != nil {
		return errors.Wrap(err, "while updating action object status")
	}

	return nil
}

func (r *ActionReconciler) executeAction(ctx context.Context, action *v1alpha1.Action) error {
	sa, err := r.svc.EnsureWorkflowSAExists(ctx, action)
	if err != nil {
		return errors.Wrap(err, "while creating runner service account")
	}

	if err := r.svc.EnsureRunnerInputDataCreated(ctx, sa.Name, action); err != nil {
		return errors.Wrap(err, "while creating runner input")
	}

	if err := r.svc.EnsureRunnerExecuted(ctx, sa.Name, action); err != nil {
		return errors.Wrap(err, "while executing runner")
	}

	action.Status = r.successStatus(action, v1alpha1.RunningActionPhase, "Kubernetes runner executed. Waiting for finish phase.")
	if err := r.Status().Update(ctx, action); err != nil {
		return errors.Wrap(err, "while updating status of executed action")
	}

	return nil
}

func (r *ActionReconciler) handleRunningAction(ctx context.Context, action *v1alpha1.Action) error {
	// if changed update and return?
	if err := r.svc.EnsureRunnerStatusIsUpToDate(ctx, action); err != nil {
		return errors.Wrap(err, "while ensuring runner status is up to date")
	}

	out, err := r.svc.GetRunnerJobStatus(ctx, action)
	if err != nil {
		return errors.Wrap(err, "while getting runner job status")
	}

	if !out.Finished {
		return nil
	}

	switch out.JobStatus {
	case batchv1.JobComplete:
		action.Status = r.successStatus(action, v1alpha1.SucceededActionPhase, "Runner finished successfully")
	case batchv1.JobFailed:
		action.Status = r.failStatus(action, v1alpha1.FailedActionPhase, "Runner finished unsuccessfully")
	default:
		action.Status = r.failStatus(action, v1alpha1.FailedActionPhase, "Unknown runner job status")
	}
	if err := r.Status().Update(ctx, action); err != nil {
		return errors.Wrap(err, "while updating status of executed action")
	}

	return nil
}

func (r *ActionReconciler) failStatus(action *v1alpha1.Action, phase v1alpha1.ActionPhase, msg string) v1alpha1.ActionStatus {
	return r.setStatus(action, corev1.EventTypeWarning, phase, msg)
}

func (r *ActionReconciler) successStatus(action *v1alpha1.Action, phase v1alpha1.ActionPhase, msg string) v1alpha1.ActionStatus {
	return r.setStatus(action, corev1.EventTypeNormal, phase, msg)
}

func (r *ActionReconciler) setStatus(action *v1alpha1.Action, eventType string, phase v1alpha1.ActionPhase, msg string) v1alpha1.ActionStatus {
	r.recorder.Event(action, eventType, string(phase), msg)

	statusCpy := action.Status.DeepCopy()
	statusCpy.Phase = phase
	statusCpy.Message = ptr.String(msg)
	statusCpy.LastTransitionTime = metav1.Now()
	statusCpy.ObservedGeneration = action.Generation

	if statusCpy.Phase == action.Status.Phase {
		statusCpy.LastTransitionTime = action.Status.LastTransitionTime
	}

	return *statusCpy
}

func (r *ActionReconciler) SetupWithManager(mgr ctrl.Manager, maxConcurrentReconciles int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Action{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: maxConcurrentReconciles,
		}).
		Owns(&batchv1.Job{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
