package manifestgen

import (
	"strings"

	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/attribute"
	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/common"
	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/implementation"
	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/interfacegen"
	"capact.io/capact/internal/cli/alpha/manifestgen"
	"github.com/pkg/errors"
	"k8s.io/utils/strings/slices"
)

type genManifestFun func(opts common.ManifestGenOptions) (map[string]string, error)

func generateAttribute(opts common.ManifestGenOptions) (map[string]string, error) {
	files, err := attribute.GenerateAttributeFile(opts)
	if err != nil {
		return nil, errors.Wrap(err, "while generating attribute file")
	}
	return files, nil
}

func generateType(opts common.ManifestGenOptions) (map[string]string, error) {
	if slices.Contains(opts.ManifestsType, common.ImplementationManifest) {
		// type files has been already generated in the implementation step
		return nil, nil
	}
	files, err := interfacegen.GenerateInterfaceFile(opts, manifestgen.GenerateTypeTemplatingConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while generating type templating config")
	}
	return files, nil
}

func generateInterfaceGroup(opts common.ManifestGenOptions) (map[string]string, error) {
	files, err := interfacegen.GenerateInterfaceFile(opts, manifestgen.GenerateInterfaceGroupTemplatingConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while generating interface group templating config")
	}
	return files, nil
}

func generateInterface(opts common.ManifestGenOptions) (map[string]string, error) {
	if slices.Contains(opts.ManifestsType, common.TypeManifest) || slices.Contains(opts.ManifestsType, common.ImplementationManifest) {
		inputPath := common.CreateManifestPath(common.TypeManifest, opts.ManifestPath) + "-input"
		opts.TypeInputPath = common.AddRevisionToPath(inputPath, opts.Revision)
	}

	if slices.Contains(opts.ManifestsType, common.TypeManifest) {
		outputsuffix := strings.Split(opts.ManifestPath, ".")
		outputPath := common.CreateManifestPath(common.TypeManifest, outputsuffix[0]) + ".config"
		opts.TypeOutputPath = common.AddRevisionToPath(outputPath, opts.Revision)
	}

	files, err := interfacegen.GenerateInterfaceFile(opts, manifestgen.GenerateInterfaceTemplatingConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while generating interface templating config")
	}
	return files, nil
}

func generateImplementation(opts common.ManifestGenOptions) (map[string]string, error) {
	files, err := implementation.GenerateImplementationManifest(opts)
	if err != nil {
		return nil, errors.Wrap(err, "while generating implementation manifest")
	}
	return files, nil
}
