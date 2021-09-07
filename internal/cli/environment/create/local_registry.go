package create

import (
	"context"
	"io/ioutil"
	"os"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/config"
	"capact.io/capact/internal/cli/printer"

	"github.com/docker/cli/cli/streams"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/go-connections/nat"
)

const (
	// ContainerRegistryImage defines local Docker registry image.
	ContainerRegistryImage = "registry:2"
	// ContainerRegistryName defines name for the container registry. It's also used as DNS for registry.
	ContainerRegistryName = "capact-registry.localhost"
	// ContainerRegistryPort defines port for the container registry.
	// TODO: allow setting different port by user
	ContainerRegistryPort = "5000"
	// ContainerRegistry defines container DNS with port.
	ContainerRegistry = ContainerRegistryName + ":" + ContainerRegistryPort
)

// LocalRegistry create a local Docker registry used to pushed locally build images.
func LocalRegistry(ctx context.Context, status *printer.Status) (err error) {
	defer func() {
		status.End(err == nil)
	}()

	dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	// Get local_registry local path
	localRegistryFolder, err := config.GetDefaultConfigPath("local_registry")
	if err != nil {
		return err
	}

	status.Step("Storing local registry data under %s", localRegistryFolder)
	if err := os.MkdirAll(localRegistryFolder, os.ModePerm); err != nil { // ensure folder exists
		return err
	}

	if err := ensureRegistryImage(ctx, dockerCli, status); err != nil {
		return err
	}

	status.Step("Creating local registry %s", ContainerRegistry)
	// Equal to:
	// docker container run -d -p 5000:5000  --restart=always  --name capact-registry.localhost  --network capact -v $HOME/.config/capact/local_registry:/var/lib/registry registry:2
	if err != nil {
		return err
	}
	cnt := &container.Config{
		Image: ContainerRegistryImage,
	}
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

	createdCnt, err := dockerCli.ContainerCreate(ctx, cnt, host, nil, nil, ContainerRegistryName)
	if err != nil {
		return err
	}

	return dockerCli.ContainerStart(ctx, createdCnt.ID, types.ContainerStartOptions{})
}

// RegistryConnWithNetwork connects container registry with a given network.
func RegistryConnWithNetwork(ctx context.Context, networkID string) error {
	//docker network connect ${networkID} capact-registry.localhost
	dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	return dockerCli.NetworkConnect(ctx, networkID, ContainerRegistryName, nil)
}

func ensureRegistryImage(ctx context.Context, dockerCli *client.Client, status *printer.Status) error {
	found, err := foundRegistryImageLocally(ctx, dockerCli)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	return pullRegistryImage(ctx, dockerCli, status)
}

func isDockerNotFoundErr(err error) bool {
	type notFound interface {
		NotFound()
	}
	_, ok := err.(notFound)
	return ok
}

func foundRegistryImageLocally(ctx context.Context, dockerCli *client.Client) (bool, error) {
	_, _, err := dockerCli.ImageInspectWithRaw(ctx, ContainerRegistryImage)
	switch {
	case err == nil: // found
		return true, nil
	case isDockerNotFoundErr(err):
		return false, nil
	default:
		return false, err
	}
}

func pullRegistryImage(ctx context.Context, dockerCli *client.Client, status *printer.Status) error {
	out := streams.NewOut(os.Stdout)
	if !cli.VerboseMode.IsTracing() {
		status.Step("Pulling %q image", ContainerRegistryImage)
		out = streams.NewOut(ioutil.Discard)
	}

	reader, err := dockerCli.ImagePull(ctx, ContainerRegistryImage, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	err = jsonmessage.DisplayJSONMessagesToStream(reader, out, nil)
	if err != nil {
		return err
	}

	return reader.Close()
}
