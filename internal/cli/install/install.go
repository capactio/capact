package install

import (
	"context"
	"io"

	"capact.io/capact/internal/cli/capact"
	"capact.io/capact/internal/cli/printer"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)

func Install(ctx context.Context, w io.Writer, k8sCfg *rest.Config, opts capact.Options) (err error) {
	status := printer.NewStatus(w, "Installing Capact on cluster...")
	defer func() {
		status.End(err == nil)
	}()

	version := opts.Parameters.Version
	if version == "@local" {
		//status.Step("Building local images")
		if len(opts.SkipImages) != 0 && len(opts.FocusImages) != 0 {
			return errors.New("can not skip and focus images at the same time")
		}

		images, err := capact.SelectImages(opts.FocusImages, opts.SkipImages)
		if err != nil {
			return errors.Wrap(err, "while selecting images")
		}

		created, err := capact.BuildImages(w, capact.LocalDockerPath, capact.LocalDockerTag, images)
		if err != nil {
			return errors.Wrap(err, "while building images")
		}

		status.Step("Loading Docker images")
		for _, image := range created {
			err = capact.LoadImage(opts.Name, image)
			if err != nil {
				return errors.Wrap(err, "while loading images into env")
			}
		}
	}

	configuration, err := capact.GetActionConfiguration(k8sCfg, opts.Namespace)
	if err != nil {
		return err
	}

	helm := capact.NewHelm(configuration, opts)

	status.Step("Applying Capact CRDs")
	err = helm.InstallCRD()
	if err != nil {
		return err
	}

	status.Step("Creating namespace %s", opts.Namespace)
	err = capact.CreateNamespace(k8sCfg, opts.Namespace)
	if err != nil {
		return err
	}

	err = helm.InstallComponnents(w, status)
	if err != nil {
		return err
	}

	return nil
}
