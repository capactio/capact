package list

import (
	"fmt"
	"time"

	"projectvoltron.dev/voltron/internal/ocftool/credstore"
	"projectvoltron.dev/voltron/pkg/httputil"
	ochclient "projectvoltron.dev/voltron/pkg/och/client/public/generated"

	"github.com/spf13/cobra"
)

func NewCmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "This command consists of multiple subcommands to interact with OCH server.",
	}

	cmd.AddCommand(NewInterface())
	return cmd
}

// TODO: move it from here
func getOCHClient(server string) (*ochclient.Client, error) {
	store := credstore.NewOCH()
	user, pass, err := store.Get(server)
	if err != nil {
		return nil, err
	}

	httpClient := httputil.NewClient(30*time.Second, false,
		httputil.WithBasicAuth(user, pass))

	return ochclient.NewClient(httpClient, fmt.Sprintf("%s/graphql", server)), nil
}

