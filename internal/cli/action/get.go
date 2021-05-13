package action

import (
	"context"
	"fmt"
	"io"
	"time"

	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/internal/k8s-engine/graphql/namespace"
	gqlengine "capact.io/capact/pkg/engine/api/graphql"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

type GetOptions struct {
	ActionName string
	Namespace  string
	Output     string
}

func Get(ctx context.Context, opts GetOptions, w io.Writer) error {
	server := config.GetDefaultContext()

	actionCli, err := client.NewCluster(server)
	if err != nil {
		return err
	}

	var acts []*gqlengine.Action

	ctxWithNs := namespace.NewContext(ctx, opts.Namespace)

	if opts.ActionName != "" {
		act, err := actionCli.GetAction(ctxWithNs, opts.ActionName)
		if err != nil {
			return err
		}

		if act == nil {
			return errors.Errorf("Action %s not found", opts.ActionName)
		}

		acts = append(acts, act)
	} else {
		acts, err = actionCli.ListActions(ctxWithNs, &gqlengine.ActionFilter{})
		if err != nil {
			return err
		}
	}

	printAction, err := selectPrinter(opts.Output)
	if err != nil {
		return err
	}

	return printAction(opts.Namespace, acts, w)
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
			time.Since(act.CreatedAt.Time).String(),
		})
	}

	table.AppendBulk(data)
	table.Render()

	return nil
}
