package manifestgen

import (
	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/attribute"
	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/common"
	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/implementation"
	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/interfacegen"
	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/typegen"
	"capact.io/capact/internal/cli/alpha/manifestgen"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/utils/strings/slices"
)

// NewCmd returns a cobra.Command for content generation operations.
func NewCmd() *cobra.Command {
	var opts common.ManifestGenOptions
	cmd := &cobra.Command{
		Use:   "manifest-gen",
		Short: "Manifests generation",
		Long:  "Subcommand for various manifest generation operations",
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return askInteractivelyForParameters(opts)
		},
	}

	cmd.AddCommand(attribute.NewAttribute())
	cmd.AddCommand(typegen.NewType())
	cmd.AddCommand(interfacegen.NewInterface())
	cmd.AddCommand(implementation.NewCmd())

	cmd.PersistentFlags().StringP("output", "o", "generated", "Path to the output directory for the generated manifests")
	cmd.PersistentFlags().Bool("overwrite", false, "Overwrite existing manifest files")

	return cmd
}

func askInteractivelyForParameters(opts common.ManifestGenOptions) error {
	var err error
	opts.ManifestsType, err = askForManifestType()
	if err != nil {
		return errors.Wrap(err, "while asking for manifest type")
	}

	opts.ManifestPath, err = common.AskForManifestPathSuffix("Manifest path suffix")
	if err != nil {
		return errors.Wrap(err, "while asking for manifest path suffix")
	}

	revision, err := common.AskForManifestRevision("Manifests revision")
	if err != nil {
		return errors.Wrap(err, "while asking for manifest revision")
	}
	opts.Revision = revision

	metadata, err := askForCommonMetadataInformation()
	if err != nil {
		return errors.Wrap(err, "while getting the common metadata information")
	}
	opts.Metadata = *metadata

	generatingManifestsFun := map[string]genManifestFn{
		common.AttributeManifest:      generateAttribute,
		common.TypeManifest:           generateType,
		common.InterfaceGroupManifest: generateInterfaceGroup,
		common.InterfaceManifest:      generateInterface,
		common.ImplementationManifest: generateImplementation,
	}
	var manifestCollection manifestgen.ManifestCollection

	for manifestType, fn := range generatingManifestsFun {
		if !slices.Contains(opts.ManifestsType, manifestType) {
			continue
		}
		manifests, err := fn(opts)
		if err != nil {
			return errors.Wrap(err, "while generating manifest file")
		}
		manifestCollection = mergeManifests(manifestCollection, manifests)
	}

	opts.Directory, err = common.AskForDirectory("path to the output directory for the generated manifests", "generated")
	if err != nil {
		return errors.Wrap(err, "while asking for output directory")
	}

	if manifestgen.DoesAnyManifestAlreadyExistInDir(manifestCollection, opts.Directory) {
		opts.Overwrite, err = askIfOverwrite()
		if err != nil {
			return errors.Wrap(err, "while asking if overwrite existing manifest files")
		}
	}

	if err := manifestgen.WriteManifestFiles(opts.Directory, manifestCollection, opts.Overwrite); err != nil {
		return errors.Wrap(err, "while writing manifest files")
	}
	return nil
}

func mergeManifests(manifestsCollections ...manifestgen.ManifestCollection) (result manifestgen.ManifestCollection) {
	result = make(manifestgen.ManifestCollection)
	for _, manifestCollection := range manifestsCollections {
		for path, content := range manifestCollection {
			result[path] = content
		}
	}
	return result
}
