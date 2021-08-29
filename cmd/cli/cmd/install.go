package cmd

import (
	"fmt"
	"os"
	"time"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/environment/create"
	"capact.io/capact/internal/cli/heredoc"

	"capact.io/capact/internal/cli/capact"

	"capact.io/capact/internal/cli/install"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// NewInstall returns a cobra.Command for installing Capact in the created env
func NewInstall() *cobra.Command {
	opts := capact.Options{}

	installCmd := &cobra.Command{
		Use:   "install [OPTIONS]",
		Short: "install Capact into a given environment",
		Long:  "Use this command to install the Capact version in the environment.",
		Example: heredoc.WithCLIName(`
			# Install latest Capact version from main branch
			<cli> install

			# Install Capact 0.1.0 version
			<cli> install --version 0.1.0

			# Install Capact from local git repository. Needs to be run from the main directory
			<cli> install --version @local`, cli.Name),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Validate(); err != nil {
				return err
			}

			k8sCfg, err := config.GetConfig()
			if err != nil {
				return errors.Wrap(err, "while creating k8s config")
			}

			return install.Install(cmd.Context(), os.Stdout, k8sCfg, opts)
		},
	}

	flags := installCmd.Flags()

	flags.StringVar(&opts.Parameters.Version, "version", capact.LatestVersionTag, "Capact version. Possible values @latest, @local, 0.3.0, ...")
	flags.StringVar(&opts.Name, "name", create.DefaultClusterName, "Cluster name, overrides config.")
	flags.StringVar(&opts.Namespace, "namespace", capact.Namespace, "Capact namespace.")
	flags.StringVar(&opts.Environment, "environment", capact.KindEnv, "Capact environment.")
	flags.StringSliceVar(&opts.InstallComponents, "install-component", capact.Components.All(), "Components names that should be installed. Takes comma-separated list.")
	flags.StringSliceVar(&opts.BuildImages, "build-image", capact.Images.All(), "Local images names that should be build when using @local version. Takes comma-separated list.")
	flags.BoolVar(&opts.Parameters.IncreaseResourceLimits, "increase-resource-limits", true, "Enables higher resource requests and limits for components.")
	flags.DurationVar(&opts.Timeout, "timeout", 10*time.Minute, `Maximum time during which the upgrade process is being watched, where "0" means "infinite". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`)
	flags.BoolVar(&opts.UpdateHostsFile, "update-hosts-file", true, "Updates /etc/hosts with entry for Capact GraphQL Gateway.")
	flags.BoolVar(&opts.UpdateTrustedCerts, "update-trusted-certs", true, "Add Capact GraphQL Gateway certificate.")
	flags.StringVar(&opts.Parameters.Override.HelmRepoURL, "helm-repo-url", capact.HelmRepoStable, fmt.Sprintf("Capact Helm chart repository URL. Use %s tag to select repository which holds the latest Helm chart versions.", capact.LatestVersionTag))
	flags.BoolVar(&opts.RegistryEnabled, "enable-registry", false, "If specified, Capact images are pushed to registry.")
	flags.StringSliceVar(&opts.Parameters.Override.CapactStringOverrides, "capact-overrides", []string{}, "Overrides for Capact component.")
	flags.StringSliceVar(&opts.Parameters.Override.IngressStringOverrides, "ingress-controller-overrides", []string{}, "Overrides for Ingress controller component.")
	flags.StringSliceVar(&opts.Parameters.Override.CertManagerStringOverrides, "cert-manager-overrides", []string{}, "Overrides for Cert Manager component.")

	return installCmd
}
