package dockerutil

import (
	"context"
	"io/ioutil"
	"os"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/printer"

	"github.com/docker/cli/cli/streams"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
)

// EnsureImage pulls a given image only if not found locally.
func EnsureImage(ctx context.Context, dockerCli *client.Client, status printer.Status, imageRef string) error {
	found, err := foundImageLocally(ctx, dockerCli, imageRef)
	if err != nil {
		return err
	}
	if found {
		status.Infof("Image %s already present", imageRef)
		return nil
	}

	return pullRegistryImage(ctx, dockerCli, status, imageRef)
}

func isDockerNotFoundErr(err error) bool {
	type notFound interface {
		NotFound()
	}
	_, ok := err.(notFound)
	return ok
}

func foundImageLocally(ctx context.Context, dockerCli *client.Client, ref string) (bool, error) {
	_, _, err := dockerCli.ImageInspectWithRaw(ctx, ref)
	switch {
	case err == nil: // found
		return true, nil
	case isDockerNotFoundErr(err):
		return false, nil
	default:
		return false, err
	}
}

func pullRegistryImage(ctx context.Context, dockerCli *client.Client, status printer.Status, ref string) (err error) {
	defer func() { status.End(err == nil) }()

	out := streams.NewOut(os.Stdout)
	if !cli.VerboseMode.IsTracing() {
		status.Step("Pulling %q image", ref)
		out = streams.NewOut(ioutil.Discard)
	}

	reader, err := dockerCli.ImagePull(ctx, ref, types.ImagePullOptions{
		All:      false,
		Platform: os.Getenv("DOCKER_DEFAULT_PLATFORM"),
	})
	if err != nil {
		return err
	}

	err = jsonmessage.DisplayJSONMessagesToStream(reader, out, nil)
	if err != nil {
		return err
	}

	return reader.Close()
}
