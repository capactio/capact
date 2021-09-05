package archiveimages

import (
	"fmt"
	"os"

	"capact.io/capact/internal/cli"
	archiveimages "capact.io/capact/internal/cli/alpha/archive-images"
	"capact.io/capact/internal/cli/capact"
	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/cli/printer"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewFromHelmCharts returns a cobra.Command to archive images from the Capact Helm charts.
func NewFromHelmCharts() *cobra.Command {
	opts := archiveimages.HelmArchiveImagesOptions{
		CapactOpts: capact.Options{
			DryRun:     true,
			Replace:    true,
			ClientOnly: true,
		},
	}
	cmd := &cobra.Command{
		Use:   "helm",
		Short: "Archive all the Docker container images used in Capact Helm charts",
		Example: heredoc.WithCLIName(`
			# Archive images from the stable Capact Helm repository from version 0.5.0
			<cli> alpha archive-images helm --version 0.5.0 --output ./capact-images-0.5.0.tar

			# Archive images from  Helm Chart released from the the '0fbf562' commit on the main branch
			<cli> alpha archive-images helm --version 0.4.0-0fbf562 --helm-repo-url @latest > ./capact-images-0.4.0-0fbf562.tar

			# You can use gzip to save the image file and make the backup smaller.
			<cli> alpha archive-images helm --version 0.5.0 --output ./capact-images-0.5.0.tar.gz --compress gzip

			# You can pipe output to use custom gzip
			<cli> alpha archive-images helm --version 0.5.0 --output-stdout | gzip > myimage_latest.tar.gz
		`, cli.Name),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if err = opts.Resolve(); err != nil {
				return errors.Wrap(err, "while resolving version")
			}

			if err = opts.Validate(); err != nil {
				return errors.Wrap(err, "while resolving version")
			}

			var status printer.Status = printer.NewNoopStatus()
			if cli.VerboseMode.IsEnabled() {
				status = printer.NewStatus(os.Stdout, "Archiving images...")
			}
			defer func() {
				status.End(err == nil)
			}()

			return archiveimages.CapactHelmCharts(cmd.Context(), status, opts)
		},
	}

	cmd.Flags().StringVar(&opts.CapactOpts.Parameters.Version, "version", capact.LatestVersionTag, "Capact version. Possible values @latest, @local, 0.3.0, ...")
	cmd.Flags().StringVar(&opts.CapactOpts.Parameters.Override.HelmRepoURL, "helm-repo-url", capact.HelmRepoStable, fmt.Sprintf("Capact Helm chart repository URL. Use %s tag to select repository which holds the latest Helm chart versions.", capact.LatestVersionTag))
	cmd.Flags().StringVarP(&opts.Output.Path, "output", "o", "", "Write output to a file, instead of standard output.")
	cmd.Flags().BoolVar(&opts.Output.ToStdout, "output-stdout", false, "Write output to a standard output, instead of file.")
	cmd.Flags().StringVar(&opts.Compress, "compress", "", "Use a given compress algorithm. Allowed values: gzip")

	_ = cmd.RegisterFlagCompletionFunc("compress", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{archiveimages.CompressGzip}, cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("version", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("helm-repo-url", cobra.NoFileCompletions)
	_ = cmd.RegisterFlagCompletionFunc("output-stdout", cobra.NoFileCompletions)

	return cmd
}
