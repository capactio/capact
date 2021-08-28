package delete

import (
	"context"
	"fmt"

	"capact.io/capact/internal/cli/environment/create"
	"capact.io/capact/internal/cli/printer"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func LocalRegistry(ctx context.Context, status *printer.Status) (err error) {
	status.Step("Removing Docker registry (if there is one)...")
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	filter := filters.Arg("name", create.ContainerRegistryName)
	cnts, err := cli.ContainerList(ctx, types.ContainerListOptions{Filters: filters.NewArgs(filter)})
	if err != nil {
		return err
	}

	switch n := len(cnts); n {
	case 0:
		return nil
	case 1:
		return cli.ContainerRemove(ctx, cnts[0].ID, types.ContainerRemoveOptions{Force: true})
	default:
		return fmt.Errorf("found %d containers, when exepected only one", n)
	}
}
