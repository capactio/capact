package create

import (
	"context"
	"io"
	"os"

	"capact.io/capact/internal/cli/printer"
	"github.com/docker/docker/api/types"

	"capact.io/capact/internal/cli/config"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	// ContainerRegistryName defines name for the container registry. It's also used as DNS for registry.
	ContainerRegistryName = "capact-registry.localhost"
	// ContainerRegistryPort defines port for the container registry.
	ContainerRegistryPort = "5000"
	// ContainerRegistry defines container DNS with port.
	ContainerRegistry = ContainerRegistryName + ":" + ContainerRegistryPort
)

// LocalRegistry create a local Docker registry used to pushed locally build images.
func LocalRegistry(ctx context.Context, w io.Writer) (err error) {
	status := printer.NewStatus(w, "")
	defer func() {
		status.End(err == nil)
	}()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	// Get local_registry local path
	localRegistryFolder, err := config.GetDefaultConfigPath("local_registry")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(localRegistryFolder, os.ModePerm); err != nil { // ensure folder exists
		return err
	}

	status.Step("Creating local registry under %s", localRegistryFolder)
	// Equal to:
	// docker container run -d -p 5000:5000  --restart=always  --name capact-registry.localhost  --network capact -v $HOME/.config/capact/local_registry:/var/lib/registry registry:2
	cnt := &container.Config{Image: "registry:2"}
	host := &container.HostConfig{
		PortBindings: nat.PortMap{
			"5000/tcp": {{HostIP: "", HostPort: ContainerRegistryPort}},
		},
		RestartPolicy: container.RestartPolicy{Name: "always"},
		Mounts: []mount.Mount{{
			Type:   mount.TypeBind,
			Source: localRegistryFolder,
			Target: "/var/lib/registry",
		}},
	}

	createdCnt, err := cli.ContainerCreate(ctx, cnt, host, nil, nil, ContainerRegistryName)
	if err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, createdCnt.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	return nil
}

// RegistryConnWithNetwork connects container registry with a given network.
func RegistryConnWithNetwork(ctx context.Context, networkID string) error {
	//docker network connect ${networkID} capact-registry.localhost
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	return cli.NetworkConnect(ctx, networkID, ContainerRegistryName, nil)
}
