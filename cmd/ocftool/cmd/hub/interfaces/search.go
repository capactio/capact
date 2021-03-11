package interfaces

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"projectvoltron.dev/voltron/internal/ocftool/client"
	"projectvoltron.dev/voltron/internal/ocftool/config"
	ochclient "projectvoltron.dev/voltron/pkg/och/client/public/generated"

	"github.com/MakeNowJust/heredoc"
	"github.com/hokaccha/go-prettyjson"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/jsonpath"
	"sigs.k8s.io/yaml"
)

type searchOptions struct {
	pathPrefix string
	output     string
}

func NewSearch() *cobra.Command {
	var opts searchOptions

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search provides the ability to search for OCH Interfaces",
		Example: heredoc.Doc(`
			#  Show all interfaces in table format
			ocftool hub interfaces search
			
			# Print path for the first entry in returned response 
			ocftool hub interfaces search -oyaml

			# Print path for the first entry in returned response 
			ocftool hub interfaces search -o=jsonpath="{.interfaces[0]['path']}"
			
			# Print paths
			ocftool hub interfaces search -o=jsonpath="{range .interfaces[*]}{.path}{'\n'}{end}"
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listInterfaces(opts, os.Stdout)
		},
	}

	flags := cmd.Flags()

	flags.StringVar(&opts.pathPrefix, "path-prefix", "cap.interface.*", "Pattern of the path of a given Interface, e.g. cap.interface.*")
	flags.StringVarP(&opts.output, "output", "o", "table", "Output format. One of:\njson|yaml|table|jsonpath=...")

	return cmd
}

func listInterfaces(opts searchOptions, w io.Writer) error {
	cli, err := client.NewHub(config.GetDefaultContext())
	if err != nil {
		return err
	}

	interfaces, err := cli.InterfacesWithPrefixFilter(context.TODO(), opts.pathPrefix)
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

type printer func(pattern string, in *ochclient.InterfacesWithPrefixFilter, w io.Writer) error

func selectPrinter(format string) (printer, error) {
	switch format {
	case "json":
		return printJSON, nil
	case "jsonpath":
		return printJSONPath, nil
	case "yaml":
		return printYAML, nil
	case "table":
		return printTable, nil
	}

	return nil, fmt.Errorf("unknow output format %q", format)
}

func printJSONPath(pattern string, in *ochclient.InterfacesWithPrefixFilter, w io.Writer) error {
	out, err := toInterface(in)
	if err != nil {
		return err
	}
	j := jsonpath.New("out")
	if err := j.Parse(pattern); err != nil {
		return err
	}

	return j.Execute(w, out)
}

func printJSON(_ string, in *ochclient.InterfacesWithPrefixFilter, w io.Writer) error {
	out, err := prettyjson.Marshal(in)
	if err != nil {
		return err
	}

	_, err = w.Write(out)
	return err
}

func printYAML(_ string, in *ochclient.InterfacesWithPrefixFilter, w io.Writer) error {
	out, err := yaml.Marshal(in)
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	return err
}

func printTable(_ string, in *ochclient.InterfacesWithPrefixFilter, w io.Writer) error {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"PATH", "LATEST REVISION"})
	table.SetBorder(false)
	table.SetColumnSeparator(" ")

	var data [][]string
	for _, i := range in.Interfaces {
		data = append(data, []string{i.Path, i.LatestRevision.Revision})
	}
	table.AppendBulk(data)

	table.Render()

	return nil
}

func toInterface(src interface{}) (interface{}, error) {
	out, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}

	var dst interface{}
	return dst, json.Unmarshal(out, &dst)
}
