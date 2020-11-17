package argo

import (
	"context"
	"fmt"

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
		ArgoWorkflowRef ArgoWorkflowRef
	}
	ArgoWorkflowRef struct {
		Name      string
		Namespace string
	}
)

type Runner struct {
	wfClient v1alpha1.ArgoprojV1alpha1Interface
	log      *zap.Logger
}

// Logger (with logger)
func NewRunner(wfClient v1alpha1.ArgoprojV1alpha1Interface) *Runner {
	return &Runner{
		wfClient: wfClient,
		//log:      log,
	}
}

func (r *Runner) Start(ctx context.Context, in runner.StartInput) (runner.StartOutput, error) {
	var wfSpec wfv1.WorkflowSpec
	if err := yaml.Unmarshal(in.Manifest, &wfSpec); err != nil {
		return runner.StartOutput{}, errors.Wrap(err, "while unmarshaling workflow spec")
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

	// only create or upsert?
	wfCreated, err := r.wfClient.Workflows(in.ExecCtx.Platform.Namespace).Create(&wf)
	if err != nil {
		return runner.StartOutput{}, errors.Wrap(err, "while creating workflow")
	}

	return runner.StartOutput{
		Status: Status{
			ArgoWorkflowRef: ArgoWorkflowRef{
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

func (r *Runner) Name() string {
	return "Argo Workflow runner"
}
