package ocftool

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/iancoleman/strcase"
	"github.com/xeipuuv/gojsonschema"
)

type ValidationResult struct {
}

type ManifestValidator interface {
	ValidateYaml(r io.Reader) (*gojsonschema.Result, error)
}

type FilesystemManifestValidator struct {
	schemaRootPath string
}

func NewFilesystemManifestValidator(schemaRootPath string) ManifestValidator {
	return &FilesystemManifestValidator{
		schemaRootPath: schemaRootPath,
	}
}

type manifestMetadata struct {
	OcfVersion string `yaml:"ocfVersion"`
	Kind       string `yaml:"kind"`
}

func getManifestMetadata(yamlBytes []byte) (*manifestMetadata, error) {
	mm := &manifestMetadata{}
	err := yaml.Unmarshal(yamlBytes, mm)
	if err != nil {
		return nil, err
	}
	return mm, nil
}

func commonSchemaLoader(dir string, metadata *manifestMetadata) (*gojsonschema.SchemaLoader, error) {
	commonDir := fmt.Sprintf("%s/%s/schema/common", dir, metadata.OcfVersion)

	sl := gojsonschema.NewSchemaLoader()
	files, err := ioutil.ReadDir(commonDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		path := fmt.Sprintf("file://%s/%s", commonDir, file.Name())
		if err := sl.AddSchemas(gojsonschema.NewReferenceLoader(path)); err != nil {
			return nil, err
		}
	}

	return sl, err
}

func rootManifestJSONLoader(dir string, metadata *manifestMetadata) gojsonschema.JSONLoader {
	filename := strcase.ToKebab(metadata.Kind)
	path := fmt.Sprintf("file://%s/%s/schema/%s.json", dir, metadata.OcfVersion, filename)
	return gojsonschema.NewReferenceLoader(path)
}

func (v *FilesystemManifestValidator) ValidateYaml(r io.Reader) (*gojsonschema.Result, error) {
	yamlBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	metadata, err := getManifestMetadata(yamlBytes)
	if err != nil {
		return nil, err
	}

	sl, err := commonSchemaLoader(v.schemaRootPath, metadata)
	if err != nil {
		return nil, err
	}

	rootLoader := rootManifestJSONLoader(v.schemaRootPath, metadata)

	schema, err := sl.Compile(rootLoader)
	if err != nil {
		panic(err)
	}

	jsonBytes, err := yaml.YAMLToJSON(yamlBytes)
	if err != nil {
		return nil, err
	}

	manifestLoader := gojsonschema.NewBytesLoader(jsonBytes)

	return schema.Validate(manifestLoader)
}
