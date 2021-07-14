package install

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"capact.io/capact/internal/cli/capact"
	"capact.io/capact/internal/cli/printer"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)

// Install installs Capact
func Install(ctx context.Context, w io.Writer, k8sCfg *rest.Config, opts capact.Options) (err error) {
	status := printer.NewStatus(w, "Installing Capact on cluster...")
	defer func() {
		status.End(err == nil)
	}()

	version := opts.Parameters.Version
	if version == "@local" {
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

	err = hideLog()
	if err != nil {
		return err
	}

	err = helm.InstallComponents(w, status)
	if err != nil {
		return err
	}

	if opts.UpdateHostsFile {
		err = capact.AddGatewayToHostsFile(status)
		if err != nil {
			return err
		}
	}

	if opts.UpdateTrustedCerts {
		err = capact.TrustSelfSigned(status)
		if err != nil {
			return err
		}
	}
	status.End(true)

	welcomeMessage(w)

	return nil
}

func welcomeMessage(w io.Writer) {
	msg := `
Capact installed successfully!

To begin working with Capact, use 'capact login' command.
To read more how to use CLI, check out the documentation on https://capact.io/docs/cli/getting-started#first-use.
`
	fmt.Fprintln(w, msg)
}

func hideLog() error {
	log.SetOutput(io.Discard)	
	return nil
}
