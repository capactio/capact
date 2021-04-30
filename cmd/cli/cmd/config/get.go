package config

import (
	"io"
	"os"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/internal/cli/credstore"
	"capact.io/capact/internal/cli/heredoc"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func NewGet() *cobra.Command {
	return &cobra.Command{
		Use:   "get-contexts",
		Short: "Lists the available Hub configuration contexts",
		Example: heredoc.WithCLIName(`
			# List all the Hub configuration contexts 
			<cli> config get-contexts
		`, cli.Name),
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
