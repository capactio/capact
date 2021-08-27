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

		if opts.Registry != "" {
			registryPath = opts.Registry
		}

		// TODO can we parallelize it?
		created, err := capact.BuildImages(ctx, status, registryPath, registryTag, opts.BuildImages)
		if err != nil {
			return errors.Wrap(err, "while building images")
		}

		if opts.Registry != "" { // push to Docker registry
			if err := capact.PushImages(ctx, status, created, opts.Registry); err != nil {
				return err
			}
		} else { // load into a given environment
			if err := capact.LoadImages(ctx, created, opts); err != nil {
				return err
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
	err = capact.CreateNamespace(ctx, k8sCfg, opts.Namespace)
	if err != nil {
		return err
	}

	log.SetOutput(io.Discard)
	err = helm.InstallComponents(ctx, w, status)
	if err != nil {
		return err
	}

	if opts.UpdateHostsFile || opts.UpdateTrustedCerts {
		status.Step("Preparing local changes")
		status.End(true)
	}

	if opts.UpdateHostsFile {
		err = capact.AddGatewayToHostsFile()
		if err != nil {
			return err
		}
	}

	if opts.UpdateTrustedCerts {
		err = capact.TrustSelfSigned()
		if err != nil {
			return err
		}
	}

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
