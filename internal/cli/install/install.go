package install

import (
	"context"
	"fmt"
	"io"
	"log"

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

	err = opts.Parameters.SetCapactValuesFromOverrides()
	if err != nil {
		return errors.Wrap(err, "while parsing capact overrides")
	}

	err = opts.Parameters.ResolveVersion()
	if err != nil {
		return errors.Wrap(err, "while resolving version")
	}

	version := opts.Parameters.Version
	if version == "@local" {
		registryPath := opts.Parameters.Override.CapactValues.Global.ContainerRegistry.Path
		registryTag := opts.Parameters.Override.CapactValues.Global.ContainerRegistry.Tag
		// TODO can we parallelize it?
		created, err := capact.BuildImages(ctx, w, registryPath, registryTag, opts.BuildImages)
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

	log.SetOutput(io.Discard)
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
