package archiveimages

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/capact"
	"capact.io/capact/internal/cli/printer"
	"capact.io/capact/internal/ctxutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/pkg/errors"
)

// CompressGzip defines gzip format name
const CompressGzip = "gzip"

var (
	image = regexp.MustCompile(`image:(.*)`)
	// Special case that we need to find image declared by flag for Prometheus instance.
	// Is hard without this: https://github.com/helm/helm/issues/7754
	prometheusReloader = regexp.MustCompile(`--prometheus-config-reloader=(.*)`)
)

// CapactHelmCharts archives images from the Capact Helm charts.
func CapactHelmCharts(ctx context.Context, status printer.Status, opts HelmArchiveImagesOptions) error {
	images, err := findImagesInHelmCharts(ctx, status, opts.CapactOpts)
	if err != nil {
		return err
	}

	status.Info("Found images in Helm charts", foundImagesInfo(images))
	responseBody, err := dockerSaveDepImages(ctx, status, images)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	out, err := selectOutputForArchive(status, opts)
	if err != nil {
		return err
	}

	writer := compressOutputIfNeeded(out, opts)
	_, err = io.Copy(writer, responseBody)
	return err
}

func selectOutputForArchive(status printer.Status, opts HelmArchiveImagesOptions) (io.Writer, error) {
	if opts.Output.ToStdout {
		return os.Stdout, nil
	}
	status.Info("Save output to %s", opts.Output.Path)
	return os.Create(filepath.Clean(opts.Output.Path))
}

func compressOutputIfNeeded(out io.Writer, opts HelmArchiveImagesOptions) io.Writer {
	switch opts.Compress {
	case CompressGzip:
		return gzip.NewWriter(out)
	default:
		return out
	}
}

func findImagesInHelmCharts(ctx context.Context, status printer.Status, opts capact.Options) (map[string]struct{}, error) {
	images := map[string]struct{}{}
	for _, component := range capact.Components {
		if ctxutil.ShouldExit(ctx) {
			return nil, ctx.Err()
		}

		status.Step("Resolving images for %s", component.Name())

		component.WithOptions(&opts)
		rel, err := component.RunInstall(opts.Parameters.Version, map[string]interface{}{})
		if err != nil {
			return nil, err
		}

		manifests := make([]string, 0, len(rel.Hooks)+1)
		manifests = append(manifests, rel.Manifest)
		for _, hook := range rel.Hooks {
			if hook == nil {
				continue
			}
			manifests = append(manifests, hook.Manifest)
		}

		for _, manifest := range manifests {
			foundImages := image.FindAllStringSubmatch(manifest, -1)
			for _, img := range foundImages {
				if len(img) != 2 {
					continue
				}
				images[sanitizeImageString(img[1])] = struct{}{}
			}
		}

		// special case that we need to find image declared by flag for Prometheus instance.
		foundImages := prometheusReloader.FindAllStringSubmatch(rel.Manifest, -1)
		for _, e := range foundImages {
			if len(e) != 2 {
				continue
			}
			images[sanitizeImageString(e[1])] = struct{}{}
		}
	}

	return images, nil
}

func sanitizeImageString(in string) string {
	s := strings.Replace(in, `"`, "", 2)
	return strings.TrimSpace(s)
}

func foundImagesInfo(images map[string]struct{}) string {
	subStep := subStepFprintfFunc(4)
	var buff strings.Builder
	for i := range images {
		subStep(&buff, i)
	}
	return buff.String()
}

func subStepFprintfFunc(indent int) func(w io.Writer, format string, a ...interface{}) {
	return func(w io.Writer, format string, a ...interface{}) {
		msg := fmt.Sprintf(format, a...)
		fmt.Fprintf(w, "%s- %s\n", strings.Repeat(" ", indent), msg)
	}
}

func dockerSaveDepImages(ctx context.Context, status printer.Status, images map[string]struct{}) (io.ReadCloser, error) {
	dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	var pulled []string
	for img := range images {
		reader, err := dockerCli.ImagePull(ctx, img, types.ImagePullOptions{
			All:      false,
			Platform: os.Getenv("DOCKER_DEFAULT_PLATFORM"),
		})
		if err != nil {
			return nil, err
		}

		switch cli.VerboseMode {
		case cli.VerboseModeTracing:
			err = jsonmessage.DisplayJSONMessagesStream(reader, os.Stdout, os.Stdout.Fd(), cli.IsSmartTerminal(os.Stdout), nil)
			if err != nil {
				return nil, err
			}
		case cli.VerboseModeSimple:
			status.Step("Pulling %s", img)
		}

		if err := reader.Close(); err != nil {
			return nil, errors.Wrap(err, "while closing reader")
		}

		pulled = append(pulled, img)
	}

	return dockerCli.ImageSave(ctx, pulled)
}
