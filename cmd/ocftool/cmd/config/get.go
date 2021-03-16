package config

import (
	"io"
	"os"

	"projectvoltron.dev/voltron/internal/ocftool"
	"projectvoltron.dev/voltron/internal/ocftool/config"
	"projectvoltron.dev/voltron/internal/ocftool/credstore"
	"projectvoltron.dev/voltron/internal/ocftool/heredoc"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func NewGet() *cobra.Command {
	return &cobra.Command{
		Use:   "get-contexts",
		Short: "Print the value of a given configuration key",
		Example: heredoc.WithCLIName(`
			# List all authorized targets 
			<cli> config get-contexts
		`, ocftool.CLIName),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return getRun(os.Stdout)
		},
	}
}

func getRun(w io.Writer) error {
	out, err := credstore.ListHubServer()
	if err != nil {
		return err
	}

	def, err := config.GetDefaultContext()
	if err != nil {
		return err
	}

	printTable(def, out, w)

	return nil
}

func printTable(defaultServer string, servers []string, w io.Writer) {
	table := tablewriter.NewWriter(w)

	table.SetHeader([]string{"SERVER", "AUTH TYPE", "DEFAULT"})
	table.SetBorder(false)
	table.SetColumnSeparator(" ")

	var data [][]string
	for _, url := range servers {
		isDefault := defaultServer == url
		data = append(data, []string{url, "Basic Auth", toString(isDefault)})
	}
	table.AppendBulk(data)
	table.Render()
}

func toString(in bool) string {
	if in {
		return "YES"
	}
	return "NO"
}
