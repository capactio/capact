package ocftool

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/ghodss/yaml"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

type ValidationResult struct {
	Errors []error
}

func NewValidationResult(errors ...error) *ValidationResult {
	return &ValidationResult{
		Errors: errors,
	}
}

func (r *ValidationResult) Valid() bool {
	return len(r.Errors) == 0
}

type ManifestValidator interface {
	ValidateFile(filepath string) *ValidationResult
}

type FilesystemManifestValidator struct {
	schemaRootPath string
	commonSchemas  map[string]*gojsonschema.SchemaLoader
	rootSchemas    map[manifestMetadata]*gojsonschema.Schema
}

func NewFilesystemManifestValidator(schemaRootPath string) ManifestValidator {
	return &FilesystemManifestValidator{
		schemaRootPath: schemaRootPath,
		commonSchemas:  map[string]*gojsonschema.SchemaLoader{},
		rootSchemas:    map[manifestMetadata]*gojsonschema.Schema{},
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

func commonSchemaLoader(dir string, ocfVersion string) (*gojsonschema.SchemaLoader, error) {
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

func rootManifestJSONLoader(dir string, metadata *manifestMetadata) gojsonschema.JSONLoader {
	filename := strcase.ToKebab(metadata.Kind)
	path := fmt.Sprintf("file://%s/%s/schema/%s.json", dir, metadata.OcfVersion, filename)
	return gojsonschema.NewReferenceLoader(path)
}

func (v *FilesystemManifestValidator) getCommonShemaLoader(ocfVersion string) (*gojsonschema.SchemaLoader, error) {
	if sl, ok := v.commonSchemas[ocfVersion]; ok {
		return sl, nil
	}

	sl, err := commonSchemaLoader(v.schemaRootPath, ocfVersion)
	if err != nil {
		return nil, err
	}

	v.commonSchemas[ocfVersion] = sl
	return sl, nil
}

func (v *FilesystemManifestValidator) getSchema(metadata *manifestMetadata) (*gojsonschema.Schema, error) {
	if schema, ok := v.rootSchemas[*metadata]; ok {
		return schema, nil
	}

	rootLoader := rootManifestJSONLoader(v.schemaRootPath, metadata)

	sl, err := v.getCommonShemaLoader(metadata.OcfVersion)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get common schema loader")
	}

	schema, err := sl.Compile(rootLoader)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to compile schema for %s/%s", metadata.OcfVersion, metadata.Kind)
	}

	v.rootSchemas[*metadata] = schema

	return schema, nil
}

func (v *FilesystemManifestValidator) validateYamlFromReader(r io.Reader) *ValidationResult {
	yamlBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return NewValidationResult(errors.Wrap(err, "failed to read data"))
	}
	metadata, err := getManifestMetadata(yamlBytes)
	if err != nil {
		return NewValidationResult(errors.Wrap(err, "failed to get manifest metadata"))
	}

	schema, err := v.getSchema(metadata)
	if err != nil {
		return NewValidationResult(errors.Wrap(err, "failed to get JSON schema"))
	}

	jsonBytes, err := yaml.YAMLToJSON(yamlBytes)
	if err != nil {
		return NewValidationResult(errors.Wrap(err, "cannot convert YAML manifest to JSON"))
	}

	manifestLoader := gojsonschema.NewBytesLoader(jsonBytes)

	jsonschemaResult, err := schema.Validate(manifestLoader)
	if err != nil {
		return NewValidationResult(errors.Wrap(err, "error occurred during JSON schema validation"))
	}

	result := NewValidationResult()

	for _, err := range jsonschemaResult.Errors() {
		result.Errors = append(result.Errors, fmt.Errorf("%v", err.String()))
	}

	return result
}

func getDummyTemplateFuncsMap() template.FuncMap {
	return template.FuncMap{
		"actionFrom": func(interface{}) string { return "" },
		"action":     func(x interface{}) interface{} { return x },
	}
}

func (v *FilesystemManifestValidator) ValidateFile(filepath string) *ValidationResult {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return NewValidationResult(err)
	}

	tmpl, err := template.New(filepath).Funcs(sprig.GenericFuncMap()).Funcs(getDummyTemplateFuncsMap()).Parse(string(data))
	if err != nil {
		return NewValidationResult(errors.Wrap(err, "failed to parse manifest template"))
	}

	buf := &bytes.Buffer{}

	if err := tmpl.Execute(buf, map[string]interface{}{}); err != nil {
		return NewValidationResult(errors.Wrap(err, "failed to render manifest template"))
	}

	templateString := strings.ReplaceAll(buf.String(), "<no value>", "")

	return v.validateYamlFromReader(bytes.NewBufferString(templateString))
}
