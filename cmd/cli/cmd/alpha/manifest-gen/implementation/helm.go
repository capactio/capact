package implementation

import (
	"strings"

	"capact.io/capact/internal/cli/alpha/manifestgen"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var helmCfg manifestgen.HelmConfig

// NewHelm returns a cobra.Command to bootstrap Helm based manifests.
func NewHelm() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "helm [MANIFEST_PATH] [HELM_CHART_PATH]",
		Short: "Generate Helm chart based manifests",
		Long:  "Generate Helm based manifests based on a Helm chart",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("accepts two arguments: [MANIFEST_PATH] [HELM_CHART_PATH]")
			}

			path := args[0]
			if !strings.HasPrefix(path, "cap.implementation.") || len(strings.Split(path, ".")) < 4 {
				return errors.New(`manifest path must be in format "cap.implementation.[PREFIX].[NAME]"`)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			helmCfg.ManifestPath = args[0]
			helmCfg.ChartName = args[1]

			files, err := manifestgen.GenerateHelmManifests(&helmCfg)
			if err != nil {
				return errors.Wrap(err, "while generating Helm manifests")
			}

			outputDir, err := cmd.Flags().GetString("output")
			if err != nil {
				return errors.Wrap(err, "while reading output flag")
			}

			overrideManifests, err := cmd.Flags().GetBool("override")
			if err != nil {
				return errors.Wrap(err, "while overriding existing manifest")
			}

			if err := manifestgen.WriteManifestFiles(outputDir, files, overrideManifests); err != nil {
				return errors.Wrap(err, "while writing manifest files")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&helmCfg.RepoURL, "repo", "r", "", "URL of the Helm repository")
	cmd.Flags().StringVarP(&helmCfg.Version, "version", "v", "", "Version of the Helm chart")

	return cmd
}
