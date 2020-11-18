package argo

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"projectvoltron.dev/voltron/pkg/runner"

	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/pkg/errors"
	"go.uber.org/zap"
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

type (
	Status struct {
		ArgoWorkflowRef WorkflowRef
	}
	WorkflowRef struct {
		Name      string
		Namespace string
	}
)

type Runner struct {
	wfClient v1alpha1.ArgoprojV1alpha1Interface
	log      *zap.Logger
}

func NewRunner(wfClient v1alpha1.ArgoprojV1alpha1Interface) *Runner {
	return &Runner{wfClient: wfClient}
}

func (r *Runner) Start(ctx context.Context, in runner.StartInput) (*runner.StartOutput, error) {
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
	wfCreated, err := r.wfClient.Workflows(in.ExecCtx.Platform.Namespace).Create(&wf)
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
				Name:      wfCreated.Name,
				Namespace: wfCreated.Namespace,
			},
		},
	}, nil
}

// WaitForCompletion waits until Argo Workflow is finished.
func (r *Runner) WaitForCompletion(ctx context.Context, in runner.WaitForCompletionInput) error {
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

		switch obj := event.Object.(type) {
		case *wfv1.Workflow:
			if !obj.Status.FinishedAt.IsZero() {
				fmt.Printf("Workflow %q %q at %v\n", obj.Name, obj.Status.Phase, obj.Status.FinishedAt)
				return true, nil
			}
		}

		return false, nil
	}

	_, err := watchtools.UntilWithSync(ctx, lw, &wfv1.Workflow{}, nil, workflowCompleted)
	return err
}

func (r *Runner) InjectLogger(log *zap.Logger) {
	r.log = log.Named("argo")
}

func (r *Runner) Name() string {
	return "Argo Workflow runner"
}
