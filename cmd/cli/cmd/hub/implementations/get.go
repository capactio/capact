package implementations

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/internal/cli/heredoc"
	cliprinter "capact.io/capact/internal/cli/printer"
	gqlpublicapi "capact.io/capact/pkg/och/api/graphql/public"

	"github.com/hokaccha/go-prettyjson"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type getOptions struct {
	implementationPaths []string
	output              string
}

func NewGet() *cobra.Command {
	var opts getOptions

	get := &cobra.Command{
		Use:   "get",
		Short: "Displays one or multiple Implementations available on the Hub server",
		Example: heredoc.WithCLIName(`
			# Show all Implementation Revisions in table format
			<cli> hub implementations get
			
			# Show "cap.implementation.gcp.cloudsql.postgresql.install" Implementation Revisions in YAML format			
			<cli> hub implementations get cap.interface.database.postgresql.install -oyaml
		`, cli.Name),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.implementationPaths = args
			return getImpl(cmd.Context(), opts, os.Stdout)
		},
	}

	flags := get.Flags()
	flags.StringVarP(&opts.output, "output", "o", "table", "Output format. One of:\njson | yaml | table")

	return get
}

func getImpl(ctx context.Context, opts getOptions, w io.Writer) error {
	server := config.GetDefaultContext()

	cli, err := client.NewHub(server)
	if err != nil {
		return err
	}

	var (
		implementationRevisions []*gqlpublicapi.ImplementationRevision
		errors                  []error
	)

	impls, err := cli.ListImplementationRevisions(ctx, nil)
	if err != nil {
		return err
	}

	if len(opts.implementationPaths) == 0 {
		implementationRevisions = impls
	} else {
		implMap := implementationSliceToMap(impls)

		for _, path := range opts.implementationPaths {
			foundImpls, found := implMap[path]
			if !found {
				errors = append(errors, errNotFound(path))
				continue
			}

			implementationRevisions = append(implementationRevisions, foundImpls...)
		}
	}

	printImplRev, err := selectPrinter(opts.output)
	if err != nil {
		return err
	}

	if err := printImplRev(implementationRevisions, w); err != nil {
		return err
	}

	cliprinter.PrintErrors(errors)
	return nil
}

func implementationSliceToMap(impls []*gqlpublicapi.ImplementationRevision) map[string][]*gqlpublicapi.ImplementationRevision {
	res := make(map[string][]*gqlpublicapi.ImplementationRevision)

	for i := range impls {
		impl := impls[i]
		res[impl.Metadata.Path] = append(res[impl.Metadata.Path], impl)
	}

	return res
}

func errNotFound(name string) error {
	return fmt.Errorf(`NotFound: Implementation "%s" not found`, name)
}

// TODO: all funcs should be extracted to `printers` package and return Printer Interface

type printer func(in []*gqlpublicapi.ImplementationRevision, w io.Writer) error

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

func printJSON(in []*gqlpublicapi.ImplementationRevision, w io.Writer) error {
	out, err := prettyjson.Marshal(in)
	if err != nil {
		return err
	}

	_, err = w.Write(out)
	return err
}

func printYAML(in []*gqlpublicapi.ImplementationRevision, w io.Writer) error {
	out, err := yaml.Marshal(in)
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	return err
}

func printTable(in []*gqlpublicapi.ImplementationRevision, w io.Writer) error {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"PATH", "REVISION", "ATTRIBUTES"})
	table.SetAutoWrapText(true)
	table.SetColumnSeparator(" ")
	table.SetBorder(false)
	table.SetRowLine(true)

	var data [][]string
	for _, item := range in {
		data = append(data, []string{
			item.Metadata.Path,
			item.Revision,
			attrNames(item.Metadata.Attributes),
		})
	}
	table.AppendBulk(data)

	table.Render()

	return nil
}

func attrNames(attrs []*gqlpublicapi.AttributeRevision) string {
	var out []string
	for _, a := range attrs {
		if a == nil || a.Metadata == nil {
			continue
		}
		out = append(out, a.Metadata.Path)
	}

	return strings.Join(out, "\n")
}
