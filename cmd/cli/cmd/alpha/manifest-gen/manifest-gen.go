package manifestgen

import (
	"fmt"
	"strings"

	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/common"
	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/implementation"
	_interface "capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/interface"
	"github.com/AlecAivazis/survey/v2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/utils/strings/slices"
)

var (
	interfaceType      = "interface"
	interfaceGroupType = "interfaceGroup"
	implementationType = "implementation"
	typeType           = "type"
)

// NewCmd returns a cobra.Command for content generation operations.
func NewCmd() *cobra.Command {
	var opts common.ManifestGenOptions
	cmd := &cobra.Command{
		Use:   "manifest-gen",
		Short: "Manifests generation",
		Long:  "Subcommand for various manifest generation operations",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				fmt.Println("To handle") //TODO
			}
			return askInteractivelyForParameters(opts)
		},
	}

	cmd.AddCommand(_interface.NewInterface())
	cmd.AddCommand(implementation.NewCmd())

	cmd.PersistentFlags().StringP("output", "o", "generated", "Path to the output directory for the generated manifests")
	cmd.PersistentFlags().Bool("overwrite", false, "Overwrite existing manifest files")

	return cmd
}

func askInteractivelyForParameters(opts common.ManifestGenOptions) error {
	opts.ManifestsType = askForManifestType()
	opts.Directory = common.AskForOutputDirectory("path to the output directory for the generated manifests", "generated")
	opts.Overwrite = askIfOverwrite()
	opts.ManifestPath = askForManifestPath()

	if slices.Contains(opts.ManifestsType, interfaceType) {
		_interface.GenerateInterfaceFile(opts)
	}

	if slices.Contains(opts.ManifestsType, interfaceGroupType) {
		_interface.GenerateInterfaceGroupFile(opts)
	}

	if slices.Contains(opts.ManifestsType, typeType) {
		_interface.GenerateTypeFile(opts)
	}

	if slices.Contains(opts.ManifestsType, implementationType) {
		implementation.HandleInteractiveSession(opts)
	}

	return nil
}

func askForManifestType() []string {
	var manifestTypes []string
	availableManifestsType := []string{implementationType, interfaceType, interfaceGroupType, typeType}
	prompt := []*survey.Question{
		{
			Prompt: &survey.MultiSelect{
				Message: "Which manifests do you want to generate:",
				Options: availableManifestsType,
			},
			Validate: survey.MinItems(1),
		},
	}
	survey.Ask(prompt, &manifestTypes)
	return manifestTypes
}

func askForManifestPath() string {
	var manifestPath string
	prompt := []*survey.Question{
		{
			Prompt: &survey.Input{
				Message: "Manifest path",
			},
			Validate: func(ans interface{}) error {
				if str, ok := ans.(string); !ok || !strings.HasPrefix(str, "cap.interface.") || len(strings.Split(str, ".")) < 4 {
					return errors.New(`manifest path must be in format "cap.interface.[PREFIX].[NAME]"`)

				}
				return nil
			},
		},
	}
	survey.Ask(prompt, &manifestPath)
	return manifestPath
}

func askIfOverwrite() bool {
	overwrite := false
	prompt := &survey.Confirm{
		Message: "Do you want to overwrite existing manifest files?",
	}
	survey.AskOne(prompt, &overwrite)
	return overwrite
}
