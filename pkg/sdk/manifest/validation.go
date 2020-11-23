package manifest

import (
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

type ValidationResult struct {
	Errors []error
}

func newValidationResult(errors ...error) ValidationResult {
	return ValidationResult{
		Errors: errors,
	}
}

func (r *ValidationResult) Valid() bool {
	return len(r.Errors) == 0
}

type Validator interface {
	ValidateFile(filepath string) (ValidationResult, error)
}

type ocfVersion string

type kind string

type loadedOCFSchema struct {
	common *gojsonschema.SchemaLoader
	kind   map[kind]*gojsonschema.Schema
}

type FilesystemManifestValidator struct {
	schemaRootPath string
	cachedSchemas  map[ocfVersion]*loadedOCFSchema
}

func NewFilesystemValidator(schemaRootPath string) Validator {
	return &FilesystemManifestValidator{
		schemaRootPath: schemaRootPath,
		cachedSchemas:  map[ocfVersion]*loadedOCFSchema{},
	}
}

type manifestMetadata struct {
	OCFVersion ocfVersion `yaml:"ocfVersion"`
	Kind       kind       `yaml:"kind"`
}

func getManifestMetadata(yamlBytes []byte) (manifestMetadata, error) {
	mm := manifestMetadata{}
	err := yaml.Unmarshal(yamlBytes, &mm)
	if err != nil {
		return mm, err
	}
	return mm, nil
}

func commonSchemaLoader(dir string, ocfVersion ocfVersion) (*gojsonschema.SchemaLoader, error) {
	commonDir := fmt.Sprintf("%s/%s/schema/common", dir, ocfVersion)

	sl := gojsonschema.NewSchemaLoader()
	files, err := ioutil.ReadDir(commonDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list common schemas directory")
	}

	for _, file := range files {
		path := fmt.Sprintf("file://%s/%s", commonDir, file.Name())
		if err := sl.AddSchemas(gojsonschema.NewReferenceLoader(path)); err != nil {
			return nil, errors.Wrapf(err, "cannot load common schema %s", path)
		}
	}

	return sl, nil
}

func rootManifestJSONLoader(dir string, metadata manifestMetadata) gojsonschema.JSONLoader {
	filename := strcase.ToKebab(string(metadata.Kind))
	path := fmt.Sprintf("file://%s/%s/schema/%s.json", dir, metadata.OCFVersion, filename)
	return gojsonschema.NewReferenceLoader(path)
}

func (v *FilesystemManifestValidator) getSchema(metadata manifestMetadata) (*gojsonschema.Schema, error) {
	var ok bool
	var cachedSchema *loadedOCFSchema

	if cachedSchema, ok = v.cachedSchemas[metadata.OCFVersion]; !ok {
		cachedSchema = &loadedOCFSchema{
			common: nil,
			kind:   map[kind]*gojsonschema.Schema{},
		}
		v.cachedSchemas[metadata.OCFVersion] = cachedSchema
	}

	if schema, ok := cachedSchema.kind[metadata.Kind]; ok {
		return schema, nil
	}

	rootLoader := rootManifestJSONLoader(v.schemaRootPath, metadata)

	if cachedSchema.common == nil {
		sl, err := commonSchemaLoader(v.schemaRootPath, metadata.OCFVersion)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get common schema loader")
		}
		cachedSchema.common = sl
	}

	schema, err := cachedSchema.common.Compile(rootLoader)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to compile schema for %s/%s", metadata.OCFVersion, metadata.Kind)
	}

	cachedSchema.kind[metadata.Kind] = schema

	return schema, nil
}

func (v *FilesystemManifestValidator) validateYamlBytes(yamlBytes []byte) (ValidationResult, error) {
	metadata, err := getManifestMetadata(yamlBytes)
	if err != nil {
		return newValidationResult(errors.Wrap(err, "failed to read manifest metadata")), err
	}

	schema, err := v.getSchema(metadata)
	if err != nil {
		return newValidationResult(), errors.Wrap(err, "failed to get JSON schema")
	}

	jsonBytes, err := yaml.YAMLToJSON(yamlBytes)
	if err != nil {
		return newValidationResult(errors.Wrap(err, "cannot convert YAML manifest to JSON")), err
	}

	manifestLoader := gojsonschema.NewBytesLoader(jsonBytes)

	jsonschemaResult, err := schema.Validate(manifestLoader)
	if err != nil {
		return newValidationResult(errors.Wrap(err, "error occurred during JSON schema validation")), err
	}

	result := newValidationResult()

	for _, err := range jsonschemaResult.Errors() {
		result.Errors = append(result.Errors, fmt.Errorf("%v", err.String()))
	}

	return result, err
}

func (v *FilesystemManifestValidator) ValidateFile(filepath string) (ValidationResult, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return newValidationResult(), err
	}

	return v.validateYamlBytes(data)
}
