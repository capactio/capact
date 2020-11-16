package argo

import (
	"context"
	"fmt"
	"github.com/Project-Voltron/voltron/cmd/argo-runner/runner"
	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	//apierrors "k8s.io/apimachinery/pkg/api/errors"
	watchtools "k8s.io/client-go/tools/watch"
	"sigs.k8s.io/yaml"
)

const (
	wfManagedByLabelKey = "runner.projectvoltron.dev/created-by"
	runnerName          = "argo-runner"
)

type Runner struct {
	wfClient v1alpha1.ArgoprojV1alpha1Interface
	log      *zap.Logger
}

func NewRunner(wfClient v1alpha1.ArgoprojV1alpha1Interface) *Runner {
	return &Runner{
		wfClient: wfClient,
		//log:      log,
	}
}

// TODO input/output as in sdk
// Logger (with logger)
func (r *Runner) Start(ctx context.Context, execCtx runner.ExecutionContext, manifest []byte) error {
	var wfSpec wfv1.WorkflowSpec
	if err := yaml.Unmarshal(manifest, &wfSpec); err != nil {
		return errors.Wrap(err, "while unmarshaling workflow spec")
	}

	wf := wfv1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      execCtx.Name,
			Namespace: execCtx.Platform.Namespace,
			Labels: map[string]string{
				wfManagedByLabelKey: runnerName,
			},
		},
		Spec: wfSpec,
	}

	// submit the hello world workflow
	_, err := r.wfClient.Workflows(execCtx.Platform.Namespace).Create(&wf)
	if err != nil {
		return errors.Wrap(err, "while creating workflow")
	}

	//r.log.Info("Argo Workflow submitted",
	//	zap.String("name", createdWf.Name),
	//	zap.String("namespace", createdWf.Namespace),
	//)

	return nil
}

// WaitForCompletion waits until Argi Workflow is done
func (r *Runner) WaitForCompletion(ctx context.Context, execCtx runner.ExecutionContext) error {
	fieldSelector := fields.OneTermEqualSelector("metadata.name", execCtx.Name).String()
	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			opts.FieldSelector = fieldSelector
			return r.wfClient.Workflows(execCtx.Platform.Namespace).List(opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			opts.FieldSelector = fieldSelector
			return r.wfClient.Workflows(execCtx.Platform.Namespace).Watch(opts)
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
				fmt.Printf("Workflow %s %s at %v\n", obj.Name, obj.Status.Phase, obj.Status.FinishedAt)
				return true, nil
			}
		}

		return false, nil
	}

	_, err := watchtools.UntilWithSync(ctx, lw, &wfv1.Workflow{}, nil, workflowCompleted)
	return err
}
