package implementations

import (
	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/common"
	"capact.io/capact/internal/cli/alpha/manifestgen"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/utils/strings/slices"
)

type implGeneratorType string

const (
	helmTool      implGeneratorType = "Helm"
	terraformTool implGeneratorType = "Terraform"
	emptyManifest implGeneratorType = "Empty"
)

type generateFn func(opts common.ManifestGenOptions) (manifestgen.ManifestCollection, error)

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

// GenerateImplementationManifest is responsible for generating implementation manifest based on tool.
func GenerateImplementationManifest(opts common.ManifestGenOptions) (manifestgen.ManifestCollection, error) {
	tool, err := askForImplementationTool()
	if err != nil {
		return nil, errors.Wrap(err, "while asking for used implementation tool")
	}

	interfacePathSuffixAndRevision := ""
	if slices.Contains(opts.ManifestsType, string(types.InterfaceManifestKind)) {
		interfacePathSuffixAndRevision = common.AddRevisionToPath(opts.ManifestPath, opts.Revision)
	} else {
		interfacePathSuffixAndRevision, err = askForInterface()
		if err != nil {
			return nil, errors.Wrap(err, "while asking for interface path")
		}
	}
	opts.InterfacePath = common.CreateManifestPath(types.InterfaceManifestKind, interfacePathSuffixAndRevision)

	license, err := askForLicense()
	if err != nil {
		return nil, errors.Wrap(err, "while asking for license")
	}
	opts.Metadata.License.Name = &license

	toolAction := map[implGeneratorType]generateFn{
		helmTool:      generateHelmManifests,
		terraformTool: generateTerraformManifests,
		emptyManifest: generateEmptyManifests,
	}

	return toolAction[implGeneratorType(tool)](opts)
}
