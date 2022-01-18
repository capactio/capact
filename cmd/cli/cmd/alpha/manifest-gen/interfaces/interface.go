package interfaces

import (
	"strings"

	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/common"
	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/alpha/manifestgen"
	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type genManifestFn func(cfg *manifestgen.InterfaceConfig) (manifestgen.ManifestCollection, error)

// NewInterface returns a cobra.Command to bootstrap new Interface manifests.
func NewInterface() *cobra.Command {
	var interfaceCfg manifestgen.InterfaceConfig

	cmd := &cobra.Command{
		Use:     "interface [PATH]",
		Aliases: []string{"iface", "interfaces"},
		Short:   "Generate new Interface-related manifests",
		Long:    "Generate new InterfaceGroup, Interface and associated Type manifests",
		Example: heredoc.WithCLIName(`
			# Generate manifests for the cap.interface.database.postgresql.install Interface
			<cli> alpha manifest-gen interface cap.interface.database.postgresql.install`, cli.Name),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("accepts one argument: [MANIFEST_PATH]")
			}

			path := args[0]
			if !strings.HasPrefix(path, "cap.interface.") || len(strings.Split(path, ".")) < 4 {
				return errors.New(`manifest path must be in format "cap.interface.[PREFIX].[NAME]"`)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			interfaceCfg.ManifestRef.Path = args[0]
			interfaceCfg.Metadata = common.GetDefaultInterfaceMetadata()

			manifests, err := manifestgen.GenerateInterfaceManifests(&interfaceCfg)
			if err != nil {
				return errors.Wrap(err, "while generating content files")
			}

			outputDir, err := cmd.Flags().GetString("output")
			if err != nil {
				return errors.Wrap(err, "while reading output flag")
			}

			overrideManifests, err := cmd.Flags().GetBool("overwrite")
			if err != nil {
				return errors.Wrap(err, "while reading overwrite flag")
			}

			if err := manifestgen.WriteManifestFiles(outputDir, manifests, overrideManifests); err != nil {
				return errors.Wrap(err, "while writing manifest files")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&interfaceCfg.ManifestRef.Revision, "revision", "r", "0.1.0", "Revision of the Interface manifest")

	return cmd
}

// GenerateInterfaceFile generates new Interface-group file based on function passed.
func GenerateInterfaceFile(opts common.ManifestGenOptions, fn genManifestFn) (manifestgen.ManifestCollection, error) {
	interfaceCfg := manifestgen.InterfaceConfig{
		Config: manifestgen.Config{
			ManifestRef: types.ManifestRef{
				Path:     common.CreateManifestPath(types.InterfaceManifestKind, opts.ManifestPath),
				Revision: opts.Revision,
			},
		},
		Metadata: types.InterfaceMetadata{
			DocumentationURL: opts.Metadata.DocumentationURL,
			SupportURL:       opts.Metadata.SupportURL,
			IconURL:          opts.Metadata.IconURL,
			Maintainers:      opts.Metadata.Maintainers,
		},
		InputTypeRef:  opts.TypeInputPath,
		OutputTypeRef: opts.TypeOutputPath,
	}
	files, err := fn(&interfaceCfg)
	if err != nil {
		return nil, errors.Wrap(err, "while generating content files")
	}
	return files, nil
}
