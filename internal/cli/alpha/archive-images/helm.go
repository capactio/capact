package archiveimages

import (
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"capact.io/capact/internal/cli/capact"
	"capact.io/capact/internal/cli/dockerutil"
	"capact.io/capact/internal/cli/printer"
	"capact.io/capact/internal/ctxutil"
	"capact.io/capact/internal/multierror"

	"github.com/docker/docker/client"
	"k8s.io/utils/strings/slices"
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
func CapactHelmCharts(ctx context.Context, status printer.Status, opts HelmArchiveImagesOptions) (err error) {
	images, err := findImagesInHelmCharts(ctx, status, opts.CapactOpts, opts.SaveComponents)
	if err != nil {
		return err
	}

	responseBody, err := dockerSaveDepImages(ctx, status, images)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	out, closeOut, err := selectOutputForArchive(status, opts)
	if err != nil {
		return err
	}
	defer func() {
		merr := multierror.Append(err, closeOut())
		err = merr.ErrorOrNil()
	}()

	writer, closeWriter := compressOutputIfNeeded(out, opts)
	defer func() {
		merr := multierror.Append(err, closeWriter())
		err = merr.ErrorOrNil()
	}()

	_, err = io.Copy(writer, responseBody)
	return err
}

func selectOutputForArchive(status printer.Status, opts HelmArchiveImagesOptions) (io.Writer, func() error, error) {
	if opts.Output.ToStdout {
		return os.Stdout, func() error { return nil }, nil
	}
	status.Step("Saving output to %s", opts.Output.Path)
	f, err := os.Create(filepath.Clean(opts.Output.Path))
	if err != nil {
		return nil, nil, err
	}
	return f, f.Close, nil
}

func compressOutputIfNeeded(out io.Writer, opts HelmArchiveImagesOptions) (io.Writer, func() error) {
	switch opts.Compress {
	case CompressGzip:
		gzip := gzip.NewWriter(out)
		return gzip, gzip.Close
	default:
		return out, func() error { return nil }
	}
}

func findImagesInHelmCharts(ctx context.Context, status printer.Status, opts capact.Options, saveComponents []string) (map[string]struct{}, error) {
	images := map[string]struct{}{}
	for _, component := range capact.Components {
		if ctxutil.ShouldExit(ctx) {
			return nil, ctx.Err()
		}

		if !slices.Contains(saveComponents, component.Name()) {
			continue
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

func dockerSaveDepImages(ctx context.Context, status printer.Status, images map[string]struct{}) (io.ReadCloser, error) {
	dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	var pulled []string
	for img := range images {
		if err := dockerutil.EnsureImage(ctx, dockerCli, status, img); err != nil {
			return nil, err
		}
		pulled = append(pulled, img)
	}

	status.Step("Triggering docker save")
	return dockerCli.ImageSave(ctx, pulled)
}
