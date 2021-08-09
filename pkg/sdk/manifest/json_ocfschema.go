package manifest

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

type loadedOCFSchema struct {
	common *gojsonschema.SchemaLoader
	kind   map[types.ManifestKind]*gojsonschema.Schema
}

// OCFSchemaValidator validates manifests using a OCF specification, which is read from a filesystem.
type OCFSchemaValidator struct {
	fs http.FileSystem

	schemaRootPath string
	cachedSchemas  map[types.OCFVersion]*loadedOCFSchema
}

// NewOCFSchemaValidator returns a new OCFSchemaValidator.
func NewOCFSchemaValidator(fs http.FileSystem, schemaRootPath string) *OCFSchemaValidator {
	return &OCFSchemaValidator{
		schemaRootPath: schemaRootPath,
		fs:             fs,
		cachedSchemas:  map[types.OCFVersion]*loadedOCFSchema{},
	}
}

// Do validates a manifest.
func (v *OCFSchemaValidator) Do(_ context.Context, metadata types.ManifestMetadata, jsonBytes []byte) (ValidationResult, error) {
	schema, err := v.getManifestSchema(metadata)
	if err != nil {
		return newValidationResult(), errors.Wrap(err, "while getting manifest JSON schema")
	}

	manifestLoader := gojsonschema.NewBytesLoader(jsonBytes)

	jsonschemaResult, err := schema.Validate(manifestLoader)
	if err != nil {
		return newValidationResult(err), nil
	}

	result := newValidationResult()

	for _, err := range jsonschemaResult.Errors() {
		result.Errors = append(result.Errors, fmt.Errorf("%v", err.String()))
	}

	return result, err
}

// Name returns validator name.
func (v *OCFSchemaValidator) Name() string {
	return "OCFSchemaValidator"
}

func (v *OCFSchemaValidator) getManifestSchema(metadata types.ManifestMetadata) (*gojsonschema.Schema, error) {
	var ok bool
	var cachedSchema *loadedOCFSchema

	if cachedSchema, ok = v.cachedSchemas[metadata.OCFVersion]; !ok {
		cachedSchema = &loadedOCFSchema{
			common: nil,
			kind:   map[types.ManifestKind]*gojsonschema.Schema{},
		}
		v.cachedSchemas[metadata.OCFVersion] = cachedSchema
	}

	if schema, ok := cachedSchema.kind[metadata.Kind]; ok {
		return schema, nil
	}

	rootLoader := v.getRootSchemaJSONLoader(metadata)

	if cachedSchema.common == nil {
		sl, err := v.getCommonSchemaLoader(metadata.OCFVersion)
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

func (v *OCFSchemaValidator) getRootSchemaJSONLoader(metadata types.ManifestMetadata) gojsonschema.JSONLoader {
	filename := strcase.ToKebab(string(metadata.Kind))
	path := fmt.Sprintf("file://%s/%s/schema/%s.json", v.schemaRootPath, metadata.OCFVersion, filename)
	return gojsonschema.NewReferenceLoaderFileSystem(path, v.fs)
}

func (v *OCFSchemaValidator) getCommonSchemaLoader(ocfVersion types.OCFVersion) (*gojsonschema.SchemaLoader, error) {
	commonDir := fmt.Sprintf("%s/%s/schema/common", v.schemaRootPath, ocfVersion)

	sl := gojsonschema.NewSchemaLoader()

	files, err := v.readDir(commonDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list common schemas directory")
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		path := fmt.Sprintf("file://%s/%s", commonDir, file.Name())
		if err := sl.AddSchemas(gojsonschema.NewReferenceLoaderFileSystem(path, v.fs)); err != nil {
			return nil, errors.Wrapf(err, "cannot load common schema %s", path)
		}
	}

	return sl, nil
}

// readDir reads the directory named by dirname and returns
// a list of directory entries sorted by filename.
func (v *OCFSchemaValidator) readDir(dirname string) ([]os.FileInfo, error) {
	f, err := v.fs.Open(dirname)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name() < list[j].Name() })
	return list, nil
}
