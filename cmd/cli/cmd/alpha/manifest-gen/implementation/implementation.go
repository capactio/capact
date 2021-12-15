package implementation

import (
	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/common"
	"capact.io/capact/internal/cli/alpha/manifestgen"
	"github.com/AlecAivazis/survey/v2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	helmType = "Helm"
	terraformType = "Terraform"
)

// NewCmd returns a cobra.Command for Implementation manifest generation operations.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "implementation",
		Aliases: []string{"impl", "implementations"},
		Short:   "Generate new Implementation manifests",
		Long:    "Generate new Implementation manifests for various tools.",
	}

	cmd.AddCommand(NewTerraform())
	cmd.AddCommand(NewHelm())

	return cmd
}

// HandleInteractiveSession is responsible for handling interactive session with user
func HandleInteractiveSession(opts common.ManifestGenOptions) error{
	tool := askForImplementationTool()
	basedToolDir := common.AskForOutputDirectory("Path to based tool template", "")
	if tool == helmType {
		var helmCfg manifestgen.HelmConfig
		helmCfg.ManifestPath = opts.ManifestPath
		helmCfg.ChartName = basedToolDir
		files, err := manifestgen.GenerateHelmManifests(&helmCfg)
		if err != nil {
			return errors.Wrap(err, "while generating Helm manifests")
		}

		if err := manifestgen.WriteManifestFiles(opts.Directory, files, opts.Overwrite); err != nil {
			return errors.Wrap(err, "while writing manifest files")
		}
	}

	if tool == terraformType {
		var tfContentCfg manifestgen.TerraformConfig
		tfContentCfg.ManifestPath = opts.ManifestPath
		tfContentCfg.ModulePath = basedToolDir
		files, err := manifestgen.GenerateTerraformManifests(&tfContentCfg)
		if err != nil {
			return errors.Wrap(err, "while generating content files")
		}

		if err := manifestgen.WriteManifestFiles(opts.Directory, files, opts.Overwrite); err != nil {
			return errors.Wrap(err, "while writing manifest files")
		}
	}

	return nil
}

func askForImplementationTool() string{
	var selectType string
	availableManifestsType := []string{helmType, terraformType}
	typePrompt := &survey.Select{
		Message: "Based on which tool do you want to generate implementation:",
		Options: availableManifestsType,
	}
	survey.AskOne(typePrompt, &selectType)
	return selectType
}
