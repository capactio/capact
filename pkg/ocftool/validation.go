package ocftool

import (
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/iancoleman/strcase"
	"github.com/xeipuuv/gojsonschema"
)

type ManifestValidator interface {
	ValidateYaml(yamlBytes []byte) (bool, []error)
}

type FilesystemManifestValidator struct {
	schemaRootPath     string
	commonSchemaLoader *gojsonschema.SchemaLoader
}

func NewFilesystemManifestValidator(schemaRootPath string) (ManifestValidator, error) {
	sl, err := commonSchemaLoader(fmt.Sprintf("%s/schema/common"))
	if err != nil {
		return nil, err
	}

	return &FilesystemManifestValidator{
		schemaRootPath:     schemaRootPath,
		commonSchemaLoader: sl,
	}, nil
}

type manifestMetadata struct {
	Kind string `yaml:"kind"`
}

func getManifestKind(yamlBytes []byte) (string, error) {
	mm := &manifestMetadata{}
	err := yaml.Unmarshal(yamlBytes, mm)
	if err != nil {
		return "", err
	}
	return mm.Kind, nil
}

func commonSchemaLoader(dir string) (*gojsonschema.SchemaLoader, error) {
	sl := gojsonschema.NewSchemaLoader()
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		path := fmt.Sprintf("file://%s/%s", dir, file)
		if err := sl.AddSchemas(gojsonschema.NewStringLoader(path)); err != nil {
			return nil, err
		}
	}

	return sl, err
}

func rootManifestJSONLoader(dir, kind string) gojsonschema.JSONLoader {
	filename := strcase.ToKebab(kind)
	path := fmt.Sprintf("file://%s/schema/%s.json", dir, filename)
	return gojsonschema.NewReferenceLoader(path)
}

func (v *FilesystemManifestValidator) ValidateYaml(yamlBytes []byte) (bool, []error) {
	kind, err := getManifestKind(yamlBytes)
	if err != nil {
		return false, []error{err}
	}

	rootLoader := rootManifestJSONLoader(v.schemaRootPath, kind)

	schema, err := v.commonSchemaLoader.Compile(rootLoader)
	if err != nil {
		panic(err)
	}

	jsonBytes, err := yaml.YAMLToJSON(yamlBytes)
	if err != nil {
		return false, []error{err}
	}

	manifestLoader := gojsonschema.NewBytesLoader(jsonBytes)

	res, err := schema.Validate(manifestLoader)
	if err != nil {
		return false, []error{err}
	}

	if !res.Valid() {
		errs := []error{}
		for _, err := range res.Errors() {
			errs = append(errs, fmt.Errorf("%s", err.String()))
		}

		return false, errs
	}

	return true, nil
}
