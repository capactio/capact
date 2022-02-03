package validation

import (
	"context"
	"encoding/json"
	"strings"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

// TypeInstanceValidation gather the necessary data to validate TypeInstance.
type TypeInstanceValidation struct {
	Alias   *string
	TypeRef types.TypeRef
	Value   interface{}
}

// ValidateTI is responsible for validating the TypeInstance.
func ValidateTI(ctx context.Context, ti *TypeInstanceValidation, hub HubClient) (Result, error) {
	if ti == nil {
		return Result{}, nil
	}

	if _, ok := ti.Value.(map[string]interface{}); !ok {
		return Result{}, nil
	}

	resultBldr := NewResultBuilder("TypeInstance value")

	typeName := getTypeNameFromPath(ti.TypeRef.Path)
	typeRevision, err := ResolveTypeRefsToJSONSchemas(ctx, hub, TypeRefCollection{
		typeName: TypeRef{
			TypeRef: ti.TypeRef,
		},
	})
	if err != nil {
		return Result{}, errors.Wrap(err, "while resolving TypeRefs to JSON Schemas")
	}

	valuesJSON, err := convertTypeInstanceValueToJSONBytes(ti.Value)
	if err != nil {
		return Result{}, errors.Wrap(err, "while converting TypeInstance value to JSON bytes")
	}

	schemaLoader := gojsonschema.NewStringLoader(typeRevision[typeName].Value)
	dataLoader := gojsonschema.NewBytesLoader(valuesJSON)

	result, err := gojsonschema.Validate(schemaLoader, dataLoader)
	if err != nil {
		return nil, err
	}
	if !result.Valid() {
		for _, err := range result.Errors() {
			name := ""
			if ti.Alias != nil {
				name = *ti.Alias
			}
			resultBldr.ReportIssue(name, err.String())
		}
	}

	return resultBldr.Result(), nil
}

func convertTypeInstanceValueToJSONBytes(values interface{}) ([]byte, error) {
	parameters := make(map[string]json.RawMessage)
	valueMap := values.(map[string]interface{})

	for name := range valueMap {
		value := valueMap[name]
		valueData, err := json.Marshal(&value)
		if err != nil {
			return nil, errors.Wrapf(err, "while marshaling %s parameter to JSON", name)
		}

		parameters[name] = valueData
	}
	return json.Marshal(parameters)
}

func getTypeNameFromPath(path string) string {
	parts := strings.Split(path, ".")
	return parts[len(parts)-1]
}
