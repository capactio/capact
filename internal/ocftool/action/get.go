package action

import (
	"context"
	"fmt"
	"io"
	"time"

	"sigs.k8s.io/yaml"

	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/namespace"
	"projectvoltron.dev/voltron/internal/ocftool/client"
	"projectvoltron.dev/voltron/internal/ocftool/config"
	gqlengine "projectvoltron.dev/voltron/pkg/engine/api/graphql"

	"github.com/AlecAivazis/survey/v2"
	"github.com/hokaccha/go-prettyjson"
	"github.com/olekukonko/tablewriter"
)

type GetOptions struct {
	ActionName string `survey:"name"`
	Namespace  string
	Output     string
}

func Get(ctx context.Context, opts GetOptions, w io.Writer) error {
	var qs []*survey.Question

	if opts.ActionName == "" {
		qs = append(qs, actionNameQuestion(""))
	}

	if opts.Namespace == "" {
		qs = append(qs, namespaceQuestion())
	}

	if err := survey.Ask(qs, &opts); err != nil {
		return err
	}

	actionCli, err := client.NewCluster(config.GetDefaultContext())
	if err != nil {
		return err
	}

	ctxWithNs := namespace.NewContext(ctx, opts.Namespace)
	act, err := actionCli.GetAction(ctxWithNs, opts.ActionName)
	if err != nil {
		return err
	}

	printAction, err := selectPrinter(opts.Output)
	if err != nil {
		return err
	}

	return printAction(opts.Namespace, act, w)
}

// TODO: all funcs should be extracted to `printers` package and return Printer Interface

type printer func(namespace string, in *gqlengine.Action, w io.Writer) error

func selectPrinter(format string) (printer, error) {
	switch format {
	case "json":
		return printJSON, nil
	case "yaml":
		return printYAML, nil
	case "table":
		return printTable, nil
	}

	return nil, fmt.Errorf("unknow output format %q", format)
}

func printJSON(_ string, in *gqlengine.Action, w io.Writer) error {
	out, err := prettyjson.Marshal(in)
	if err != nil {
		return err
	}

	_, err = w.Write(out)
	return err
}

func printYAML(_ string, in *gqlengine.Action, w io.Writer) error {
	out, err := yaml.Marshal(in)
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	return err
}

func printTable(namespace string, in *gqlengine.Action, w io.Writer) error {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"NAMESPACE", "NAME", "PATH", "RUN", "STATUS", "AGE"})
	table.SetBorder(false)
	table.SetColumnSeparator(" ")

	data := []string{
		namespace,
		in.Name,
		in.ActionRef.Path,
		toString(in.Run),
		string(in.Status.Phase),
		time.Since(in.CreatedAt.Time).String(),
	}

	table.Append(data)
	table.Render()

	return nil
}
