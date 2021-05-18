package policy

import (
	"context"
	"fmt"
	"io"

	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
)

type GetOptions struct {
	Output string
}

func Get(ctx context.Context, opts GetOptions, w io.Writer) error {
	server := config.GetDefaultContext()

	engineCli, err := client.NewCluster(server)
	if err != nil {
		return err
	}

	printPolicy, err := selectPrinter(opts.Output)
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

	return printPolicy(policy, w)
}
