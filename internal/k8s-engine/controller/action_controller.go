package controller

import (
	"context"
	"encoding/json"

	"projectvoltron.dev/voltron/internal/ptr"
	"projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
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
	k8sCli     client.Client
	log        logr.Logger
	svc        *ActionService // TODO interface
	recorder   record.EventRecorder
	implGetter OCHImplementationGetter
}

type OCHImplementationGetter interface {
	GetImplementationLatestRevision(ctx context.Context, path string) (*ochgraphql.ImplementationRevision, error)
}

// NewActionReconciler returns the ActionReconciler instance.
func NewActionReconciler(log logr.Logger, implementationGetter OCHImplementationGetter, svc *ActionService) *ActionReconciler {
	return &ActionReconciler{
		log:        log.WithName("controllers").WithName("Action"),
		svc:        svc,
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

	action := &v1alpha1.Action{}
	if err := r.k8sCli.Get(ctx, req.NamespacedName, action); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "while fetching Action CR")
		return ctrl.Result{}, err
	}

	reportOnError := func(err error, context string) (ctrl.Result, error) {
		r.recorder.Event(action, corev1.EventTypeWarning, context, err.Error())
		log.Error(err, context)
		return ctrl.Result{}, err
	}

	if action.IsBeingDeleted() {
		// TODO: currently cannot reach this state.
		// Add finalizer and handle deletion properly (cancel running actions, remove ArgoWorkflows)
		return ctrl.Result{}, nil
	}

	if action.IsUninitialized() {
		action.Status = r.successStatus(action, v1alpha1.BeingRenderedActionPhase, "Rendering runner action")
		if err := r.k8sCli.Status().Update(ctx, action); err != nil {
			return reportOnError(err, "Init runner")
		}
		return ctrl.Result{Requeue: true}, nil
	}

	if action.IsBeingRendered() {
		log.Info("Rendering runner action")
		if err := r.renderAction(ctx, action); err != nil {
			return reportOnError(err, "Render runner action")
		}
		return ctrl.Result{Requeue: true}, nil
	}

	if action.IsApprovedForExecution() {
		log.Info("Execute runner")
		result, err := r.executeAction(ctx, action)
		if err != nil {
			return reportOnError(err, "Execute runner")
		}
		return result, nil
	}

	if action.IsExecuted() {
		log.Info("Check executed runner")
		result, err := r.handleRunningAction(ctx, action)
		if err != nil {
			return reportOnError(err, "Check executed runner")
		}
		return result, nil
	}

	return ctrl.Result{}, nil
}

// renderAction renders a given action.
// Requeue for rendering nested levels (do not block to reconcile loop to render whole action at once).
// If finally rendered, sets status to v1alpha1.ReadyToRunActionPhase phase.
//
// TODO: add support for v1alpha1.AdvancedModeRenderingIterationActionPhase phase
func (r *ActionReconciler) renderAction(ctx context.Context, action *v1alpha1.Action) error {
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
		action.Status.Rendering = &v1alpha1.RenderingStatus{}
	}

	action.Status.Rendering.Action = &runtime.RawExtension{
		Raw: actionBytes,
	}

	action.Status = r.successStatus(action, v1alpha1.ReadyToRunActionPhase, "Runner action is rendered and ready to be executed")
	if err := r.k8sCli.Status().Update(ctx, action); err != nil {
		return errors.Wrap(err, "while updating action object status")
	}

	return nil
}

// executeAction executes action (run, dryRun, cancel etc) and set v1alpha1.RunningActionPhase.
//
// TODO: add support v1alpha1.BeingCancelledActionPhase phase.
func (r *ActionReconciler) executeAction(ctx context.Context, action *v1alpha1.Action) (ctrl.Result, error) {
	sa, err := r.svc.EnsureWorkflowSAExists(ctx, action)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "while creating runner service account")
	}

	if err := r.svc.EnsureRunnerInputDataCreated(ctx, sa.Name, action); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "while creating runner input")
	}

	if err := r.svc.EnsureRunnerExecuted(ctx, sa.Name, action); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "while executing runner")
	}

	action.Status = r.successStatus(action, v1alpha1.RunningActionPhase, "Kubernetes runner executed. Waiting for finish phase.")
	if err := r.k8sCli.Status().Update(ctx, action); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "while updating status of executed action")
	}

	// requeue for checking execution status
	return ctrl.Result{Requeue: true}, nil
}

// handleRunningAction checks execution status. If completed, sets final state v1alpha1.SucceededActionPhase,
// v1alpha1.CancelledActionPhase, or v1alpha1.FailedActionPhase depends on currently scheduled activity.
//
// TODO: add support for cancel phase.
func (r *ActionReconciler) handleRunningAction(ctx context.Context, action *v1alpha1.Action) (ctrl.Result, error) {
	type newStatusCreator func(ctx context.Context, action *v1alpha1.Action) (*v1alpha1.ActionStatus, error)
	steps := []newStatusCreator{
		r.reportedRunnerStatus,
		r.runnerJobStatus,
	}

	for _, step := range steps {
		newStatus, err := step(ctx, action)
		if err != nil {
			return ctrl.Result{}, err
		}
		if newStatus == nil {
			continue
		}

		action.Status = *newStatus
		if err := r.k8sCli.Status().Update(ctx, action); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "while updating status of executed action")
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// status didn't change, no need to requeue
	return ctrl.Result{}, nil
}

func (r *ActionReconciler) reportedRunnerStatus(ctx context.Context, action *v1alpha1.Action) (*v1alpha1.ActionStatus, error) {
	reportedStatus, err := r.svc.GetReportedRunnerStatus(ctx, action)
	if err != nil {
		return nil, errors.Wrap(err, "while getting scheduled runner status")
	}

	if !reportedStatus.Changed {
		return nil, nil
	}

	statusCpy := action.Status.DeepCopy()
	if statusCpy.Runner == nil {
		statusCpy.Runner = &v1alpha1.RunnerStatus{
			Interface: "why.we.need.that.?", // TODO: Any thoughts Paweł?
		}
	}
	statusCpy.Runner.Status = &runtime.RawExtension{
		Raw: reportedStatus.Status,
	}

	return statusCpy, nil
}

func (r *ActionReconciler) runnerJobStatus(ctx context.Context, action *v1alpha1.Action) (*v1alpha1.ActionStatus, error) {
	out, err := r.svc.GetRunnerJobStatus(ctx, action)
	if err != nil {
		return nil, errors.Wrap(err, "while getting runner job status")
	}

	if !out.Finished {
		return nil, nil
	}

	var outStatus v1alpha1.ActionStatus
	switch out.JobStatus {
	case batchv1.JobComplete:
		outStatus = r.successStatus(action, v1alpha1.SucceededActionPhase, "Runner finished successfully")
	case batchv1.JobFailed:
		outStatus = r.failStatus(action, v1alpha1.FailedActionPhase, "Runner finished unsuccessfully")
	default:
		outStatus = r.failStatus(action, v1alpha1.FailedActionPhase, "Unknown runner job status")
	}

	return &outStatus, nil
}

// failStatus sets generic status fields to indicated action failed state. Emits proper K8s Event.
func (r *ActionReconciler) failStatus(action *v1alpha1.Action, phase v1alpha1.ActionPhase, msg string) v1alpha1.ActionStatus {
	return r.newStatusForAction(action, corev1.EventTypeWarning, phase, msg)
}

// successStatus sets generic status fields to indicated action success state. Emits proper K8s Event.
func (r *ActionReconciler) successStatus(action *v1alpha1.Action, phase v1alpha1.ActionPhase, msg string) v1alpha1.ActionStatus {
	return r.newStatusForAction(action, corev1.EventTypeNormal, phase, msg)
}

func (r *ActionReconciler) newStatusForAction(action *v1alpha1.Action, eventType string, phase v1alpha1.ActionPhase, msg string) v1alpha1.ActionStatus {
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
	r.k8sCli = mgr.GetClient()
	r.recorder = mgr.GetEventRecorderFor("action-controller")

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Action{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: maxConcurrentReconciles,
		}).
		Owns(&batchv1.Job{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
