package argo

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"time"

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
)

type Runner struct {
	wfClient v1alpha1.ArgoprojV1alpha1Interface
	log      *zap.Logger
}

func NewRunner(log *zap.Logger, wfClient v1alpha1.ArgoprojV1alpha1Interface) *Runner {
	return &Runner{
		wfClient: wfClient,
		log:      log,
	}
}

func (r *Runner) Execute(manifest []byte) error {
	// todo (constructor?)
	sch := runtime.NewScheme()
	err := wfv1.AddToScheme(sch)
	if err != nil {
		return errors.Wrap(err, "while registering Workflow scheme")
	}

	wf := &wfv1.Workflow{}
	deserializer := serializer.NewCodecFactory(sch).UniversalDeserializer()
	err = runtime.DecodeInto(deserializer, manifest, wf)

	// submit the hello world workflow
	createdWf, err := r.wfClient.Workflows("ns").Create(wf)
	if err != nil {
		return errors.Wrap(err, "while creating workflow")
	}

	log := r.log.With(zap.String("name", createdWf.Name))

	log.Info("Workflow submitted")

	return r.WaitForCompletion(createdWf.Name, createdWf.Namespace, time.Second)
}

// waitForTestSuite watches the given test suite until the exitCondition is true
func (r *Runner) WaitForCompletion(name, namespace string, timeout time.Duration) error {
	ctx, cancel := watchtools.ContextWithOptionalTimeout(context.Background(), timeout)
	defer cancel()

	fieldSelector := fields.OneTermEqualSelector("metadata.name", name).String()
	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			opts.FieldSelector = fieldSelector
			return r.wfClient.Workflows(namespace).List(opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			opts.FieldSelector = fieldSelector
			return r.wfClient.Workflows(namespace).Watch(opts)
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
