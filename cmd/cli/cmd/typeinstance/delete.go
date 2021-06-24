package typeinstance

import (
	"context"
	"io"
	"os"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/internal/cli/heredoc"
	cliprinter "capact.io/capact/internal/cli/printer"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// NewDelete returns a cobra.Command for deleting a TypeInstance on a Local Hub.
func NewDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete TYPE_INSTANCE_ID...",
		Short: "Delete a given TypeInstance(s)",
		Example: heredoc.WithCLIName(`
			# Delete TypeInstances with IDs 'c49b' and '4793'
			<cli> typeinstance delete c49b 4793
		`, cli.Name),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteTI(cmd.Context(), args, os.Stdout)
		},
	}

	return cmd
}

func deleteTI(ctx context.Context, ids []string, w io.Writer) error {
	server := config.GetDefaultContext()

	hubCli, err := client.NewHub(server)
	if err != nil {
		return err
	}

	okCheck := color.New(color.FgGreen).FprintfFunc()

	var errs []error
	for _, id := range ids {
		err := hubCli.DeleteTypeInstance(ctx, id)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		okCheck(w, "TypeInstance %s deleted successfully\n", id)
	}

	cliprinter.PrintErrors(errs)

	return nil
}
