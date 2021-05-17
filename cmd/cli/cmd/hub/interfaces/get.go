package interfaces

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
	gqlpublicapi "capact.io/capact/pkg/och/api/graphql/public"

	"github.com/hokaccha/go-prettyjson"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type getOptions struct {
	interfacePaths []string
	output         string
}

var (
	allPathPrefix = "cap.interface.*"
)

func NewGet() *cobra.Command {
	var opts getOptions

	get := &cobra.Command{
		Use:   "get",
		Short: "Displays one or multiple Interfaces available on the Hub server",
		Example: heredoc.WithCLIName(`
			# Show all Interfaces in table format:
			<cli> hub interfaces get
			
			# Show "cap.interface.database.postgresql.install" Interface in JSON format:
			<cli> hub interfaces get cap.interface.database.postgresql.install -ojson
		`, cli.Name),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.interfacePaths = args
			return listInterfaces(cmd.Context(), opts, os.Stdout)
		},
	}

	flags := get.Flags()
	flags.StringVarP(&opts.output, "output", "o", "table", "Output format. One of:\njson | yaml | table")

	return get
}

func listInterfaces(ctx context.Context, opts getOptions, w io.Writer) error {
	server := config.GetDefaultContext()

	cli, err := client.NewHub(server)
	if err != nil {
		return err
	}

	var interfaces []*gqlpublicapi.Interface

	ifaces, err := cli.ListInterfacesWithLatestRevision(ctx, gqlpublicapi.InterfaceFilter{
		PathPattern: &allPathPrefix,
	})
	if err != nil {
		return err
	}

	if len(opts.interfacePaths) == 0 {
		interfaces = ifaces
	} else {
		for _, name := range opts.interfacePaths {
			iface := findInterface(ifaces, name)
			if iface == nil {
				continue
			}

			interfaces = append(interfaces, iface)
		}
	}

	format, pattern := extractOutputFormat(opts.output)
	printInterfaces, err := selectPrinter(format)
	if err != nil {
		return err
	}

	return printInterfaces(pattern, interfaces, w)
}

func findInterface(ifaces []*gqlpublicapi.Interface, name string) *gqlpublicapi.Interface {
	for i := range ifaces {
		if ifaces[i].Path == name {
			return ifaces[i]
		}
	}

	return nil
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

	return nil, fmt.Errorf("Unknown output format %q", format)
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
