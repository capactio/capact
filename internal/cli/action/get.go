package action

import (
	"context"
	"fmt"
	"io"
	"time"

	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	cliprinter "capact.io/capact/internal/cli/printer"
	"capact.io/capact/internal/k8s-engine/graphql/namespace"
	gqlengine "capact.io/capact/pkg/engine/api/graphql"
	"k8s.io/apimachinery/pkg/util/duration"

	"github.com/olekukonko/tablewriter"
)

type GetOptions struct {
	ActionNames []string
	Namespace   string
	Output      string
}

func Get(ctx context.Context, opts GetOptions, w io.Writer) error {
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

	printAction, err := selectPrinter(opts.Output)
	if err != nil {
		return err
	}

	if err := printAction(opts.Namespace, actions, w); err != nil {
		return err
	}

	cliprinter.PrintErrors(errors)
	return nil
}

func errNotFound(name string) error {
	return fmt.Errorf(`NotFound: Action "%s" not found`, name)
}

// TODO: all funcs should be extracted to `printers` package and return Printer Interface

type printer func(namespace string, in []*gqlengine.Action, w io.Writer) error

func selectPrinter(format string) (printer, error) {
	switch format {
	case "json":
		return func(_ string, in []*gqlengine.Action, w io.Writer) error {
			return printJSON(in, w)
		}, nil
	case "yaml":
		return func(_ string, in []*gqlengine.Action, w io.Writer) error {
			return printYAML(in, w)
		}, nil
	case "table":
		return printGetTable, nil
	}

	return nil, fmt.Errorf("Unknown output format %q", format)
}

func printGetTable(namespace string, in []*gqlengine.Action, w io.Writer) error {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"NAMESPACE", "NAME", "PATH", "RUN", "STATUS", "AGE"})
	table.SetBorder(false)
	table.SetColumnSeparator(" ")

	var data [][]string

	for _, act := range in {
		data = append(data, []string{
			namespace,
			act.Name,
			act.ActionRef.Path,
			toString(act.Run),
			string(act.Status.Phase),
			duration.HumanDuration(time.Since(act.CreatedAt.Time)),
		})
	}

	table.AppendBulk(data)
	table.Render()

	return nil
}
