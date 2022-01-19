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

// NewInterfaceGroup returns a cobra.Command to bootstrap new InterfaceGroup manifests.
func NewInterfaceGroup() *cobra.Command {
	var interfaceGroupCfg manifestgen.InterfaceGroupConfig

	cmd := &cobra.Command{
		Use:     "interfacegroup [PATH]",
		Aliases: []string{"igroup", "interfacegroups"},
		Short:   "Generate new InterfaceGroup manifest",
		Example: heredoc.WithCLIName(`
			# Generate manifests for the cap.interface.database.postgresql InterfaceGroup
			<cli> alpha manifest-gen interfacegroup cap.interface.database.postgresql`, cli.Name),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("accepts one argument: [MANIFEST_PATH]")
			}

			path := args[0]
			if !strings.HasPrefix(path, "cap.interface.") || len(strings.Split(path, ".")) < 3 {
				return errors.New(`manifest path must be in format "cap.interface.[NAME]"`)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			interfaceGroupCfg.ManifestRef.Path = args[0]
			interfaceGroupCfg.Metadata = common.GetDefaultInterfaceMetadata()

			manifests, err := manifestgen.GenerateInterfaceGroupTemplatingConfig(&interfaceGroupCfg)
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

	cmd.Flags().StringVarP(&interfaceGroupCfg.ManifestRef.Revision, "revision", "r", "0.1.0", "Revision of the InterfaceGroup manifest")

	return cmd
}

// GenerateInterfaceGroupFile generates new InterfaceGroup file.
func GenerateInterfaceGroupFile(opts common.ManifestGenOptions) (manifestgen.ManifestCollection, error) {
	interfaceCfg := manifestgen.InterfaceGroupConfig{
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
	}
	files, err := manifestgen.GenerateInterfaceGroupTemplatingConfig(&interfaceCfg)
	if err != nil {
		return nil, errors.Wrap(err, "while generating content files")
	}
	return files, nil
}

// GenerateInterfaceGroupFileFromInterfacePath generates new InterfaceGroup file based on Interface path.
func GenerateInterfaceGroupFileFromInterfacePath(opts common.ManifestGenOptions) (manifestgen.ManifestCollection, error) {
	interfaceCfg := manifestgen.InterfaceGroupConfig{
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
	}
	files, err := manifestgen.GenerateInterfaceGroupTemplatingConfigFromInterfacePath(&interfaceCfg)
	if err != nil {
		return nil, errors.Wrap(err, "while generating content files")
	}
	return files, nil
}
