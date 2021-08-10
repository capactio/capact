package dbpopulator

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"capact.io/capact/pkg/sdk/manifest"
	"github.com/pkg/errors"
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
		metadata, err := manifest.GetMetadata(content)
		if err != nil {
			return manifests, errors.Wrapf(err, "while unmarshaling manifest content from path %s", path)
		}

		list, ok := manifests[string(metadata.Kind)]
		if !ok {
			return nil, fmt.Errorf("Unknown manifest kind: %s", metadata.Kind)
		}
		manifests[string(metadata.Kind)] = append(list, path)
	}
	return manifests, nil
}
