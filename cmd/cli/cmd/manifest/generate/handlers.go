package generate

import (
	"strings"

	"capact.io/capact/cmd/cli/cmd/manifest/generate/attributes"
	"capact.io/capact/cmd/cli/cmd/manifest/generate/common"
	"capact.io/capact/cmd/cli/cmd/manifest/generate/implementations"
	"capact.io/capact/cmd/cli/cmd/manifest/generate/interfaces"
	"capact.io/capact/internal/cli/manifestgen"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
	"k8s.io/utils/strings/slices"
)

type genManifestFn func(opts common.ManifestGenOptions) (manifestgen.ManifestCollection, error)

func generateAttribute(opts common.ManifestGenOptions) (manifestgen.ManifestCollection, error) {
	files, err := attributes.GenerateAttributeFile(opts)
	if err != nil {
		return nil, errors.Wrap(err, "while generating attribute file")
	}
	return files, nil
}

func generateType(opts common.ManifestGenOptions) (manifestgen.ManifestCollection, error) {
	if slices.Contains(opts.ManifestsKinds, string(types.InterfaceManifestKind)) {
		inputTypeManifest, err := interfaces.GenerateInterfaceFile(opts, manifestgen.GenerateInputTypeTemplatingConfig)
		if err != nil {
			return nil, errors.Wrap(err, "while generating input type templating config")
		}
		outputTypeManifest, err := interfaces.GenerateInterfaceFile(opts, manifestgen.GenerateOutputTypeTemplatingConfig)
		if err != nil {
			return nil, errors.Wrap(err, "while generating output type templating config")
		}
		inputTypeManifest = mergeManifests(inputTypeManifest, outputTypeManifest)
		return inputTypeManifest, nil
	}
	files, err := interfaces.GenerateInterfaceFile(opts, manifestgen.GenerateTypeTemplatingConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while generating type templating config")
	}
	return files, nil
}

func generateInterfaceGroup(opts common.ManifestGenOptions) (manifestgen.ManifestCollection, error) {
	var files manifestgen.ManifestCollection
	var err error
	if slices.Contains(opts.ManifestsKinds, string(types.InterfaceManifestKind)) {
		files, err = interfaces.GenerateInterfaceGroupFileFromInterfacePath(opts)
	} else {
		files, err = interfaces.GenerateInterfaceGroupFile(opts)
	}

	if err != nil {
		return nil, errors.Wrap(err, "while generating interface group templating config")
	}
	return files, nil
}

func generateInterface(opts common.ManifestGenOptions) (manifestgen.ManifestCollection, error) {
	if slices.Contains(opts.ManifestsKinds, string(types.TypeManifestKind)) {
		opts.TypeInputPath = types.ManifestRef{
			Path:     common.CreateManifestPath(types.TypeManifestKind, opts.ManifestPath) + "-input",
			Revision: opts.Revision,
		}
		outputsuffix := strings.Split(opts.ManifestPath, ".")
		pathWithoutLastName := strings.Join(outputsuffix[:len(outputsuffix)-1], ".")
		opts.TypeOutputPath = types.ManifestRef{
			Path:     common.CreateManifestPath(types.TypeManifestKind, pathWithoutLastName) + ".config",
			Revision: opts.Revision,
		}
	}

	files, err := interfaces.GenerateInterfaceFile(opts, manifestgen.GenerateInterfaceTemplatingConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while generating interface templating config")
	}
	return files, nil
}

func generateImplementation(opts common.ManifestGenOptions) (manifestgen.ManifestCollection, error) {
	files, err := implementations.GenerateImplementationManifest(opts)
	if err != nil {
		return nil, errors.Wrap(err, "while generating implementation manifest")
	}
	return files, nil
}
