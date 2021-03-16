package interfaces

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"projectvoltron.dev/voltron/internal/ocftool"
	"projectvoltron.dev/voltron/internal/ocftool/client"
	"projectvoltron.dev/voltron/internal/ocftool/config"
	"projectvoltron.dev/voltron/internal/ocftool/heredoc"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"

	"github.com/hokaccha/go-prettyjson"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type searchOptions struct {
	pathPattern string
	output      string
}

func NewSearch() *cobra.Command {
	var opts searchOptions

	search := &cobra.Command{
		Use:   "search",
		Short: "Search provides the ability to search for OCH Interfaces",
		Example: heredoc.WithCLIName(`
			#  Show all interfaces in table format
			<cli> hub interfaces search
			
			# Print path for the first entry in returned response 
			<cli> hub interfaces search -oyaml
		`, ocftool.CLIName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listInterfaces(cmd.Context(), opts, os.Stdout)
		},
	}

	flags := search.Flags()

	flags.StringVar(&opts.pathPattern, "path-pattern", "cap.interface.*", "Pattern of the path of a given Interface, e.g. cap.interface.*")
	flags.StringVarP(&opts.output, "output", "o", "table", "Output format. One of:\njson|yaml|table")

	return search
}

func listInterfaces(ctx context.Context, opts searchOptions, w io.Writer) error {
	server, err := config.GetDefaultContext()
	if err != nil {
		return err
	}

	cli, err := client.NewHub(server)
	if err != nil {
		return err
	}

	interfaces, err := cli.ListInterfacesWithLatest(ctx, gqlpublicapi.InterfaceFilter{
		PathPattern: &opts.pathPattern,
	})
	if err != nil {
		return err
	}

	format, pattern := extractOutputFormat(opts.output)
	printInterfaces, err := selectPrinter(format)
	if err != nil {
		return err
	}

	return printInterfaces(pattern, interfaces, w)
}

func extractOutputFormat(output string) (format string, pattern string) {
	split := strings.SplitN(output, "=", 2)
	if len(split) == 1 {
		return output, ""
	}
	return split[0], split[1]
}

// TODO: all funcs should be extracted to `printers` package and return Printer Interface

type printer func(pattern string, in []*gqlpublicapi.Interface, w io.Writer) error

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

func printJSON(_ string, in []*gqlpublicapi.Interface, w io.Writer) error {
	out, err := prettyjson.Marshal(in)
	if err != nil {
		return err
	}

	_, err = w.Write(out)
	return err
}

func printYAML(_ string, in []*gqlpublicapi.Interface, w io.Writer) error {
	out, err := yaml.Marshal(in)
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	return err
}

func printTable(_ string, in []*gqlpublicapi.Interface, w io.Writer) error {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"PATH", "LATEST REVISION", "IMPLEMENTATIONS"})
	table.SetAutoWrapText(true)
	table.SetColumnSeparator(" ")
	table.SetBorder(false)
	table.SetRowLine(true)

	var data [][]string
	for _, i := range in {
		data = append(data, []string{
			i.Path,
			i.LatestRevision.Revision,
			implList(i.LatestRevision.ImplementationRevisions)},
		)
	}

	table.AppendBulk(data)
	table.Render()

	return nil
}

func implList(revisions []*gqlpublicapi.ImplementationRevision) string {
	var out []string
	for _, r := range revisions {
		if r == nil || r.Metadata == nil {
			continue
		}
		out = append(out, r.Metadata.Path)
	}
	return strings.Join(out, "\n")
}
