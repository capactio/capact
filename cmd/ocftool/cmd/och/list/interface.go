package list

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"projectvoltron.dev/voltron/cmd/ocftool/cmd/action"

	"projectvoltron.dev/voltron/internal/ocftool/config"
	ochclient "projectvoltron.dev/voltron/pkg/och/client/public/generated"

	"github.com/AlecAivazis/survey/v2"
	"github.com/hokaccha/go-prettyjson"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/jsonpath"
	"sigs.k8s.io/yaml"
)

type interfaceListOptions struct {
	pathPrefix  string
	output      string
	interactive bool
}

func NewInterface() *cobra.Command {
	var opts interfaceListOptions

	cmd := &cobra.Command{
		Use:   "interfaces",
		Short: "List OCH Interfaces",
		Example: heredoc.Doc(`
			#  List all interfaces in table format
			ocftool och list interfaces
			
			# Print path for the first entry in returned response 
			ocftool och list interfaces -o=jsonpath="{.interfaces[0]['path']}"
			
			# Print paths
			ocftool och list interfaces -o=jsonpath="{range .interfaces[*]}{.path}{'\n'}{end}"

			# Start interactive mode
			ocftool och list interfaces -i
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listInterfaces(opts, os.Stdout)
		},
	}

	flags := cmd.Flags()

	flags.StringVar(&opts.pathPrefix, "path-prefix", "cap.interface.*", "Pattern of the path of a given Interface, e.g. cap.interface.*")
	flags.StringVarP(&opts.output, "output", "o", "table", "Output format. One of:\njson|yaml|table|jsonpath=...")
	flags.BoolVarP(&opts.interactive, "interactive", "i", false, "Start interactive mode")

	return cmd
}

func listInterfaces(opts interfaceListOptions, w io.Writer) error {
	cli, err := getOCHClient(config.GetDefaultContext())
	if err != nil {
		return err
	}

	interfaces, err := cli.InterfacesWithPrefixFilter(context.TODO(), opts.pathPrefix)
	if err != nil {
		return err
	}

	if opts.interactive {
		return interactiveSelection(interfaces)
	}

	format, pattern := extractOutputFormat(opts.output)
	printInterfaces := selectPrinter(format)

	return printInterfaces(pattern, interfaces, w)
}

func interactiveSelection(in *ochclient.InterfacesWithPrefixFilter) error {
	interfaceName := ""
	prompt := &survey.Select{
		Message:  "Choose interface to run:",
		PageSize: 20,
	}
	for _, i := range in.Interfaces {
		prompt.Options = append(prompt.Options, i.Path)
	}

	if err := survey.AskOne(prompt, &interfaceName); err != nil {
		return err
	}

	opts := action.CreateOptions{
		InterfaceName: interfaceName,
		DryRun:        false,
	}
	return action.Create(context.TODO(), opts, os.Stdout)
}

func extractOutputFormat(output string) (format string, pattern string) {
	split := strings.SplitN(output, "=", 2)
	if len(split) == 1 {
		return output, ""
	}
	return split[0], split[1]
}

type printer func(pattern string, in *ochclient.InterfacesWithPrefixFilter, w io.Writer) error

// TODO: all funcs should be extracted to `printers` package and return Printer Interface

func selectPrinter(format string) printer {
	switch format {
	case "json":
		return printJSON
	case "jsonpath":
		return printJSONPath
	case "yaml":
		return printYAML
	case "table":
		return printTable
	default:
		return emptyPrinter
	}
}

func emptyPrinter(_ string, _ *ochclient.InterfacesWithPrefixFilter, _ io.Writer) error {
	return nil
}

func printJSONPath(pattern string, in *ochclient.InterfacesWithPrefixFilter, w io.Writer) error {
	out, err := toMapInterface(in)
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
	table.SetHeader([]string{"NAME", "LATEST REVISION"})
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

func toMapInterface(src interface{}) (map[string]interface{}, error) {
	out, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}

	var dst map[string]interface{}
	return dst, json.Unmarshal(out, &dst)
}
