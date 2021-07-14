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
		status.End(true)
		err = capact.AddGatewayToHostsFile(w)
		if err != nil {
			return err
		}
	}

	if opts.UpdateTrustedCerts {
		status.End(true)
		err = capact.TrustSelfSigned(w)
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

You can now use it with:

 capact login https://gateway.capact.local -u graphql -p t0p_s3cr3t
 capact typeinstance get

Check out https://capact.io/docs/introduction to see what to do next.
`
	fmt.Fprintln(w, msg)
}

func hideLog() error {
	null := "/dev/null"
	f, err := os.Open(null)
	if err != nil {
		return err
	}
	log.SetOutput(f)
	return nil
}
