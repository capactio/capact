// build +ocfmanifests

package manifests

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

// TestManifestsValid in the future will be removed and replaced with
// an `ocftool validate` command executed against all examples.
// TODO: Remove as a part of https://cshark.atlassian.net/browse/SV-21

const ocfPathPrefix = "../ocf-spec/0.0.1/schema"
const yamlFileRegex = ".yaml$"

type filenameFilter struct {
	Exact    map[string]struct{}
	Excluded map[string]struct{}
}

func TestManifestsValid(t *testing.T) {
	// Load the common schemas. Currently, the https $ref is not working as we didn't publish the spec yet.
	sl := gojsonschema.NewSchemaLoader()

	schemaRefPaths := []string{
		fmt.Sprintf("%s/common/json-schema-type.json", ocfPathPrefix),
		fmt.Sprintf("%s/common/metadata.json", ocfPathPrefix),
		fmt.Sprintf("%s/common/metadata-tags.json", ocfPathPrefix),
	}
	err := loadCommonSchemas(sl, schemaRefPaths)
	require.NoError(t, err, "while loading common schemas")

	tests := map[string]struct {
		jsonSchemaPath      string
		manifestDirectories []string
		filter              filenameFilter
	}{
		"Type manifests should be valid": {
			jsonSchemaPath: fmt.Sprintf("%s/type.json", ocfPathPrefix),
			manifestDirectories: []string{
				"core/type",
				"type",
			},
		},
		"Tag manifests should be valid": {
			jsonSchemaPath: fmt.Sprintf("%s/tag.json", ocfPathPrefix),
			manifestDirectories: []string{
				"core/tag",
				"tag",
			},
		},
		"Vendor manifests should be valid": {
			jsonSchemaPath: fmt.Sprintf("%s/vendor.json", ocfPathPrefix),
			manifestDirectories: []string{
				"vendor",
			},
		},
		"RepoMetadata manifests should be valid": {
			jsonSchemaPath: fmt.Sprintf("%s/repo-metadata.json", ocfPathPrefix),
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
			jsonSchemaPath: fmt.Sprintf("%s/interface-group.json", ocfPathPrefix),
			manifestDirectories: []string{
				"core/interface",
				"interface",
			},
			filter: filenameFilter{
				Exact: map[string]struct{}{
					"METADATA.yaml": {},
				},
			},
		},
		"Interface manifests should be valid": {
			jsonSchemaPath: fmt.Sprintf("%s/interface.json", ocfPathPrefix),
			manifestDirectories: []string{
				"core/interface",
				"interface",
			},
			filter: filenameFilter{
				Excluded: map[string]struct{}{
					"METADATA.yaml": {},
				},
			},
		},
		"Implementation manifests should be valid": {
			jsonSchemaPath: fmt.Sprintf("%s/implementation.json", ocfPathPrefix),
			manifestDirectories: []string{
				"implementation",
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			schemaLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", tc.jsonSchemaPath))
			schema, err := sl.Compile(schemaLoader)
			require.NoError(t, err, "while creating schema validator")

			// when
			for _, dirPath := range tc.manifestDirectories {
				err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return errors.Wrap(err, "while loading file/dir")
					}

					ok, err := shouldReadFile(info, tc.filter)
					if err != nil {
						return errors.Wrap(err, "while loading file/dir")
					}
					if !ok {
						return nil
					}

					manifest, err := documentLoader(path)
					if err != nil {
						return errors.Wrapf(err, "while loading manifest from path '%s'", path)
					}

					result, err := schema.Validate(manifest)
					if err != nil {
						return errors.Wrap(err, "while validating object against JSON Schema")
					}

					// then
					assertResultIsValid(t, path, result)
					return nil
				})
				require.NoError(t, err)
			}
		})
	}
}

func documentLoader(path string) (gojsonschema.JSONLoader, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	obj := map[string]interface{}{}
	if err := yaml.Unmarshal(buf, &obj); err != nil {
		return nil, err
	}

	return gojsonschema.NewGoLoader(obj), nil
}

func assertResultIsValid(t *testing.T, path string, result *gojsonschema.Result) {
	t.Helper()

	if !assert.True(t, result.Valid()) {
		t.Errorf("%s: The document is not valid. see errors:\n", path)
		for _, desc := range result.Errors() {
			t.Errorf("- %s\n", desc.String())
		}
	}
}

func loadCommonSchemas(schemaLoader *gojsonschema.SchemaLoader, paths []string) error {
	for _, path := range paths {
		jsonLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", path))
		err := schemaLoader.AddSchemas(jsonLoader)
		if err != nil {
			return err
		}
	}

	return nil
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
