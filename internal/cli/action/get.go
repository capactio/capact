package action

import (
	"context"
	"fmt"
	"time"

	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	cliprinter "capact.io/capact/internal/cli/printer"
	"capact.io/capact/internal/k8s-engine/graphql/namespace"
	gqlengine "capact.io/capact/pkg/engine/api/graphql"

	"k8s.io/apimachinery/pkg/util/duration"
)

// GetOptions holds configuration for fetching Actions.
type GetOptions struct {
	ActionNames []string
	Namespace   string
	Output      string
}

// GetOutput defines output for Get function.
type GetOutput struct {
	Actions   []*gqlengine.Action
	Namespace string
}

// Get fetches given Actions and use printer to display them in requested format.
func Get(ctx context.Context, opts GetOptions, printer *cliprinter.ResourcePrinter) error {
	server := config.GetDefaultContext()

	actionCli, err := client.NewCluster(server)
	if err != nil {
		return err
	}

	var (
		actions []*gqlengine.Action
		errors  []error
	)

	ctxWithNs := namespace.NewContext(ctx, opts.Namespace)

	if len(opts.ActionNames) == 0 {
		acts, err := actionCli.ListActions(ctxWithNs, &gqlengine.ActionFilter{})
		if err != nil {
			return err
		}

		actions = acts
	} else {
		for _, name := range opts.ActionNames {
			act, err := actionCli.GetAction(ctxWithNs, name)
			if err != nil {
				return err
			}

			if act == nil {
				errors = append(errors, errNotFound(name))
				continue
			}

			actions = append(actions, act)
		}
	}

	cliprinter.PrintErrors(errors)
	return printer.Print(GetOutput{
		Actions:   actions,
		Namespace: opts.Namespace,
	})
}

func errNotFound(name string) error {
	return fmt.Errorf(`NotFound: Action "%s" not found`, name)
}

// TableDataOnGet returns table data with Action specific properties.
func TableDataOnGet(in interface{}) (cliprinter.TableData, error) {
	out := cliprinter.TableData{}

	getOut, ok := in.(GetOutput)
	if !ok {
		return cliprinter.TableData{}, fmt.Errorf("got unexpected input type, expected action.GetOutput, got %T", in)
	}

	out.Headers = []string{"NAMESPACE", "NAME", "PATH", "RUN", "STATUS", "AGE"}
	for _, act := range getOut.Actions {
		out.MultipleRows = append(out.MultipleRows, []string{
			getOut.Namespace,
			act.Name,
			act.ActionRef.Path,
			toString(act.Run),
			string(act.Status.Phase),
			duration.HumanDuration(time.Since(act.CreatedAt.Time)),
		})
	}

	return out, nil
}
