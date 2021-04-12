package action

import (
	"context"
	"fmt"
	"io"
	"time"

	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/namespace"
	"projectvoltron.dev/voltron/internal/ocftool/client"
	"projectvoltron.dev/voltron/internal/ocftool/config"
	gqlengine "projectvoltron.dev/voltron/pkg/engine/api/graphql"

	"github.com/olekukonko/tablewriter"
)

type SearchOptions struct {
	Namespace string
	Output    string
}

func Search(ctx context.Context, opts SearchOptions, w io.Writer) error {
	server, err := config.GetDefaultContext()
	if err != nil {
		return err
	}

	actionCli, err := client.NewCluster(server)
	if err != nil {
		return err
	}

	ctxWithNs := namespace.NewContext(ctx, opts.Namespace)
	acts, err := actionCli.ListActions(ctxWithNs, &gqlengine.ActionFilter{})
	if err != nil {
		return err
	}

	printAction, err := selectListPrinter(opts.Output)
	if err != nil {
		return err
	}

	return printAction(opts.Namespace, acts, w)
}

// TODO: all funcs should be extracted to `printers` package and return Printer Interface

type listPrinter func(namespace string, in []*gqlengine.Action, w io.Writer) error

func selectListPrinter(format string) (listPrinter, error) {
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
		return printListTable, nil
	}

	return nil, fmt.Errorf("Unknown output format %q", format)
}

func printListTable(namespace string, in []*gqlengine.Action, w io.Writer) error {
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
