package policy

import (
	"context"
	"fmt"

	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/internal/cli/printer"
)

// Get current Capact Policy.
func Get(ctx context.Context, printer *printer.ResourcePrinter) error {
	server := config.GetDefaultContext()

	engineCli, err := client.NewCluster(server)
	if err != nil {
		return err
	}

	policy, err := engineCli.GetPolicy(ctx)
	if err != nil {
		return err
	}

	if policy == nil {
		return fmt.Errorf("Policy is empty")
	}

	return printer.Print(policy)
}
