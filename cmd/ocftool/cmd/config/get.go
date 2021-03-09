package config

import (
	"io"
	"os"
	"projectvoltron.dev/voltron/internal/ocftool/config"

	"projectvoltron.dev/voltron/internal/ocftool/credstore"

	"github.com/MakeNowJust/heredoc"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func NewGet() *cobra.Command {
	return &cobra.Command{
		Use:   "get-contexts",
		Short: "Print the value of a given configuration key",
		Example: heredoc.Doc(`
			$ ocftool config get-contexts
		`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return getRun(os.Stdout)
		},
	}
}

func getRun(w io.Writer) error {
	store := credstore.NewOCH()
	out, err := store.List()

	if err != nil {
		return err
	}

	printTable(out, w)

	return nil
}

func printTable(in map[string]string, w io.Writer) {
	table := tablewriter.NewWriter(w)

	table.SetHeader([]string{"SERVER", "USERNAME", "AUTH TYPE", "DEFAULT"})
	table.SetBorder(false)
	table.SetColumnSeparator(" ")

	def := config.GetDefaultContext()

	var data [][]string
	for url, user := range in {
		isDefault := def == url
		data = append(data, []string{url, user, "Basic Auth", toString(isDefault)})
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
