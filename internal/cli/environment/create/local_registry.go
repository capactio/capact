package create

import (
	"capact.io/capact/internal/cli/printer"
	"context"
	"github.com/docker/docker/api/types"
	"io"
	"os"

	"capact.io/capact/internal/cli/config"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const containerRegistryName = "capact-registry.localhost"

func LocalRegistry(ctx context.Context, w io.Writer) (id string, err error) {
	status := printer.NewStatus(w, "")
	defer func() {
		status.End(err == nil)
	}()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}

	// Get local_registry local path
	localRegistryFolder, err := config.GetDefaultConfigPath("local_registry")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(localRegistryFolder, os.ModePerm); err != nil { // ensure folder exists
		return "", err
	}
	status.Step("Creating local registry under %s...", localRegistryFolder)
	// Equal to:
	// docker container run -d \
	//  -p 5000:5000 \
	//  --restart=always \
	//  --name capact-registry.localhost \
	//  --network capact \
	//  -v $HOME/.config/capact/local_registry:/var/lib/registry \
	//  registry:2
	cnt := &container.Config{Image: "registry:2"}
	host := &container.HostConfig{
		PortBindings: nat.PortMap{
			"5000/tcp": {{HostIP: "", HostPort: "5000"}},
		},
		RestartPolicy: container.RestartPolicy{Name: "always"},
		Mounts: []mount.Mount{{
			Type:   mount.TypeBind,
			Source: localRegistryFolder,
			Target: "/var/lib/registry",
		}},
	}

	createdCnt, err := cli.ContainerCreate(ctx, cnt, host, nil, nil, "capact-registry.localhost")
	if err != nil {
		return "", err
	}
	if err := cli.ContainerStart(ctx, createdCnt.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}
	return createdCnt.ID, nil
}

func RegistryConnWithNetwork(ctx context.Context, networkID string) error {
	//docker network connect networkID capact-registry.localhost
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	return cli.NetworkConnect(ctx, networkID, containerRegistryName, nil)
}
