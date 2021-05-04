// build +ocfmanifests

package manifests

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"capact.io/capact/internal/cli/schema"
	"capact.io/capact/pkg/sdk/manifest"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

const ocfPathPrefix = "../ocf-spec"
const yamlFileRegex = ".yaml$"

var interfaceGroupFilenamesFilter = map[string]struct{}{
	"postgresql.yaml": {},
	"jira.yaml":       {},
	"argo.yaml":       {},
	"cloudsql.yaml":   {},
	"helm.yaml":       {},
	"generic.yaml":    {},
}

type filenameFilter struct {
	Exact    map[string]struct{}
	Excluded map[string]struct{}
}

func TestManifestsValid(t *testing.T) {
	validator := manifest.NewFilesystemValidator(&schema.LocalFileSystem{}, ocfPathPrefix)

	tests := map[string]struct {
		jsonSchemaPath      string
		manifestDirectories []string
		filter              filenameFilter
	}{
		"Type manifests should be valid": {
			jsonSchemaPath: "type.json",
			manifestDirectories: []string{
				"core/type",
				"type",
			},
		},
		"Attribute manifests should be valid": {
			jsonSchemaPath: "attribute.json",
			manifestDirectories: []string{
				"core/attribute",
				"attribute",
			},
		},
		"Vendor manifests should be valid": {
			jsonSchemaPath: "vendor.json",
			manifestDirectories: []string{
				"vendor",
			},
		},
		"RepoMetadata manifests should be valid": {
			jsonSchemaPath: "repo-metadata.json",
			manifestDirectories: []string{
				"core",
			},
			filter: filenameFilter{
				Exact: map[string]struct{}{
					"repo-metadata.yaml": {},
				},
			},
		},
		"InterfaceGroup manifests should be valid": {
			jsonSchemaPath: "interface-group.json",
			manifestDirectories: []string{
				"core/interface",
				"interface",
			},
			filter: filenameFilter{
				Exact: interfaceGroupFilenamesFilter,
			},
		},
		"Interface manifests should be valid": {
			jsonSchemaPath: "interface.json",
			manifestDirectories: []string{
				"core/interface",
				"interface",
			},
			filter: filenameFilter{
				Excluded: interfaceGroupFilenamesFilter,
			},
		},
		"Implementation manifests should be valid": {
			jsonSchemaPath: "implementation.json",
			manifestDirectories: []string{
				"implementation",
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given

			// when
			for _, dirPath := range tc.manifestDirectories {
				err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return errors.Wrap(err, "iterating over files/dirs")
					}

					ok, err := shouldReadFile(info, tc.filter)
					if err != nil {
						return errors.Wrap(err, "while loading file/dir")
					}
					if !ok {
						return nil
					}

					result, err := validator.ValidateFile(path)

					// then

					require.Nil(t, err, "returned error: %v", err)
					require.True(t, result.Valid(), "%s is not valid, errors: %v", path, result.Errors)
					return nil
				})
				require.NoError(t, err)
			}
		})
	}
}

func shouldReadFile(fileInfo os.FileInfo, filter filenameFilter) (bool, error) {
	if fileInfo.IsDir() {
		return false, nil
	}

	filename := fileInfo.Name()
	if filter.Exact != nil {
		_, exists := filter.Exact[filename]
		if !exists {
			return false, nil
		}

		return true, nil
	}

	if filter.Excluded != nil {
		_, exists := filter.Excluded[filename]
		if exists {
			return false, nil
		}
	}

	matched, err := regexp.MatchString(yamlFileRegex, filename)
	if err != nil {
		return false, errors.Wrap(err, "while matching regex")
	}

	return matched, nil
}
