package argo

import (
	"context"
	"fmt"

	"projectvoltron.dev/voltron/pkg/runner"

	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"
	"sigs.k8s.io/yaml"
)

const (
	wfManagedByLabelKey = "runner.projectvoltron.dev/created-by"
	runnerName          = "argo-runner"
)

// Provides info to easily identify started Argo Workflow.
type (
	Status struct {
		ArgoWorkflowRef WorkflowRef `json:"argoWorkflowRef"`
	}
	WorkflowRef struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	}
)

var _ runner.ActionRunner = &Runner{}

// Runner provides functionality to run and wait for Argo Workflow.
type Runner struct {
	wfClient v1alpha1.ArgoprojV1alpha1Interface
	log      *zap.Logger
}

// NewRunner returns new instance of Argo Runner.
func NewRunner(wfClient v1alpha1.ArgoprojV1alpha1Interface) *Runner {
	return &Runner{wfClient: wfClient}
}

// Start the Argo Workflow from the given manifest.
func (r *Runner) Start(ctx context.Context, in runner.StartInput) (*runner.StartOutput, error) {
	if in.ExecCtx.DryRun {
		return nil, errors.New("DryRun support not implemented")
	}
	var wfSpec wfv1.WorkflowSpec
	if err := yaml.Unmarshal(in.Manifest, &wfSpec); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling workflow spec")
	}

	if wfSpec.ServiceAccountName == "" {
		wfSpec.ServiceAccountName = in.ExecCtx.Platform.ServiceAccountName
	}

	wf := wfv1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      in.ExecCtx.Name,
			Namespace: in.ExecCtx.Platform.Namespace,
			Labels: map[string]string{
				wfManagedByLabelKey: runnerName,
			},
		},
		Spec: wfSpec,
	}

	// TODO: how should we handle retries?
	// * create and ignore already exist error [currently implemented]
	// * implement upsert. But workflow can be already in running phase, so we can mess up it.
	// * return error if already exists.
	// * create and if already exits then cancel workflow and rerun it.
	_, err := r.wfClient.Workflows(in.ExecCtx.Platform.Namespace).Create(&wf)
	switch {
	case err == nil:
	case apierrors.IsAlreadyExists(err):
		r.log.Info("ArgoWorkflow already exists. Skip create/update action.",
			zap.String("name", wf.Name),
			zap.String("namespace", wf.Namespace),
		)
	default:
		return nil, errors.Wrap(err, "while updating Argo Workflow")
	}

	return &runner.StartOutput{
		Status: Status{
			ArgoWorkflowRef: WorkflowRef{
				Name:      wf.Name,
				Namespace: wf.Namespace,
			},
		},
	}, nil
}

// WaitForCompletion waits until Argo Workflow is finished.
func (r *Runner) WaitForCompletion(ctx context.Context, in runner.WaitForCompletionInput) (*runner.WaitForCompletionOutput, error) {
	fieldSelector := fields.OneTermEqualSelector("metadata.name", in.ExecCtx.Name).String()
	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			opts.FieldSelector = fieldSelector
			return r.wfClient.Workflows(in.ExecCtx.Platform.Namespace).List(opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			opts.FieldSelector = fieldSelector
			return r.wfClient.Workflows(in.ExecCtx.Platform.Namespace).Watch(opts)
		},
	}

	workflowCompleted := func(event watch.Event) (bool, error) {
		switch event.Type {
		case watch.Modified, watch.Added:
		case watch.Deleted:
			// We need to abort to avoid cases of recreation and not to silently watch the wrong (new) object
			return false, fmt.Errorf("workflow was deleted")
		default:
			return false, nil
		}

		status, _ := statusFromEvent(&event)
		if !status.FinishedAt.IsZero() {
			return true, nil
		}

		return false, nil
	}

	lastEvent, err := watchtools.UntilWithSync(ctx, lw, &wfv1.Workflow{}, nil, workflowCompleted)
	if err != nil {
		return nil, err
	}

	status, err := statusFromEvent(lastEvent)
	if err != nil {
		return nil, err
	}

	return &runner.WaitForCompletionOutput{
		FinishedSuccessfully: status.Phase == wfv1.NodeSucceeded,
		Message:              status.Message,
	}, nil
}

func statusFromEvent(event *watch.Event) (wfv1.WorkflowStatus, error) {
	if event == nil {
		return wfv1.WorkflowStatus{}, errors.New("got nil event")
	}
	switch obj := event.Object.(type) {
	case *wfv1.Workflow:
		return obj.Status, nil
	default:
		return wfv1.WorkflowStatus{}, errors.Errorf("Wrong event object, want *wfv1.Workflow got %T", obj)
	}
}

// InjectLogger requests logger injection.
func (r *Runner) InjectLogger(log *zap.Logger) {
	r.log = log.Named("argo")
}

// Name returns runner name.
func (r *Runner) Name() string {
	return "argo.workflow"
}
