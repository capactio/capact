package dbpopulator

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

// Order in which manifests will be loaded into DB
var ordered = []string{
	"Attribute",
	"Type",
	"InterfaceGroup",
	"Interface",
	"Implementation",
	"RepoMetadata",
	"Vendor",
}

type manifestMetadata struct {
	OCFVersion string `yaml:"ocfVersion"`
	Kind       string `yaml:"kind"`
}

func getManifestMetadata(yamlBytes []byte) (manifestMetadata, error) {
	mm := manifestMetadata{}
	err := yaml.Unmarshal(yamlBytes, &mm)
	return mm, err
}

// Group returns a map of the provided Manifests Paths, grouped by the manifest kind.
func Group(paths []string) (map[string][]string, error) {
	manifests := map[string][]string{}
	for _, kind := range ordered {
		manifests[kind] = []string{}
	}
	for _, path := range paths {
		// may just read first 3 lines if there are performance issues
		content, err := ioutil.ReadFile(filepath.Clean(path))
		if err != nil {
			return manifests, errors.Wrapf(err, "while reading file from path %s", path)
		}
		metadata, err := getManifestMetadata(content)
		if err != nil {
			return manifests, errors.Wrapf(err, "while unmarshaling manifest content from path %s", path)
		}

		list, ok := manifests[metadata.Kind]
		if !ok {
			return nil, fmt.Errorf("Unknown manifest kind: %s", metadata.Kind)
		}
		manifests[metadata.Kind] = append(list, path)
	}
	return manifests, nil
}
