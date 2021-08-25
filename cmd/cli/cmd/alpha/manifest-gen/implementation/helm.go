package implementation

import (
	"strings"

	"capact.io/capact/internal/cli/alpha/manifestgen"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewHelm returns a cobra.Command to bootstrap Helm based manifests.
func NewHelm() *cobra.Command {
	var helmCfg manifestgen.HelmConfig

	cmd := &cobra.Command{
		Use:   "helm [MANIFEST_PATH] [HELM_CHART_NAME]",
		Short: "Generate Helm chart based manifests",
		Long:  "Generate Implementation manifests based on a Helm chart",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("accepts two arguments: [MANIFEST_PATH] [HELM_CHART_NAME]")
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

			overrideManifests, err := cmd.Flags().GetBool("overwrite")
			if err != nil {
				return errors.Wrap(err, "while reading overwrite flag")
			}

			if err := manifestgen.WriteManifestFiles(outputDir, files, overrideManifests); err != nil {
				return errors.Wrap(err, "while writing manifest files")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&helmCfg.InterfacePathWithRevision, "interface", "i", "", "Path with revision of the Interface, which is implemented by this Implementation")
	cmd.Flags().StringVarP(&helmCfg.ManifestRevision, "revision", "r", "0.1.0", "Revision of the Implementation manifest")
	cmd.Flags().StringVar(&helmCfg.ChartRepoURL, "repo", "", "URL of the Helm repository")
	cmd.Flags().StringVarP(&helmCfg.ChartVersion, "version", "v", "", "Version of the Helm chart")

	return cmd
}
