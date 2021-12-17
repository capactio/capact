package implementation

import (
	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/common"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/utils/strings/slices"
)

var (
	helmTool      = "Helm"
	terraformTool = "Terraform"
	emptyManifest = "Empty"
)

type generateFun func(opts common.ManifestGenOptions) (map[string]string, error)

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
	cmd.AddCommand(NewEmpty())

	return cmd
}

// GenerateImplementationManifest is responsible for generating implementation manifest based on tool
func GenerateImplementationManifest(opts common.ManifestGenOptions) (map[string]string, error) {
	tool, err := askForImplementationTool()
	if err != nil {
		return nil, errors.Wrap(err, "while asking for used implementation tool")
	}

	interfacePathSuffixAndRevision := ""
	if slices.Contains(opts.ManifestsType, common.InterfaceManifest) {
		interfacePathSuffixAndRevision = common.AddRevisionToPath(opts.ManifestPath, opts.Revision)
	} else {
		interfacePathSuffixAndRevision, err = askForInterface()
		if err != nil {
			return nil, errors.Wrap(err, "while asking for interface path")
		}
	}
	opts.InterfacePath = common.CreateManifestPath(common.InterfaceManifest, interfacePathSuffixAndRevision)

	license, err := askForLicense()
	if err != nil {
		return nil, errors.Wrap(err, "while asking for license")
	}
	opts.Metadata.License = license

	toolAction := map[string]generateFun{
		helmTool:      generateHelmManifests,
		terraformTool: generateTerraformManifests,
		emptyManifest: generateEmptyManifests,
	}

	return toolAction[tool](opts)
}
