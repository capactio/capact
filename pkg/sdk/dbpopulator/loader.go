package dbpopulator

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
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
	if err != nil {
		return mm, err
	}
	return mm, nil
}

func Group(paths []string) (map[string][]string, error) {
	manifests := map[string][]string{}
	for _, kind := range ordered {
		manifests[kind] = []string{}
	}
	for _, path := range paths {
		// may just read first 3 lines if there are performance issues
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return manifests, err
		}
		metadata, err := getManifestMetadata(content)
		if err != nil {
			return manifests, err
		}

		list, ok := manifests[metadata.Kind]
		if !ok {
			return nil, fmt.Errorf("Unknow manifest kind: %s", metadata.Kind)
		}
		manifests[metadata.Kind] = append(list, path)

	}
	return manifests, nil
}

func List(ochPath string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(ochPath, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") {
			files = append(files, currentPath)
		}
		return nil
	})

	if err != nil {
		return files, err
	}
	return files, nil
}
