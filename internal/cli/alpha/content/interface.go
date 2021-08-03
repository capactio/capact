package content

import (
	"fmt"

	"capact.io/capact/pkg/sdk/manifest"
	"github.com/pkg/errors"
)

// InterfaceConfig stores the input parameters for Interface content generation
type InterfaceConfig struct {
	Config
}

// GenerateInterfaceManifests generates manifest files for a new Interface.
func GenerateInterfaceManifests(cfg *InterfaceConfig) (map[string]string, error) {
	cfgs := []*templatingConfig{
		{
			Template: typeManifestTemplate,
			Input: templatingInput{
				Name:   cfg.ManifestName,
				Prefix: cfg.ManifestsPrefix,
			},
		},
		{
			Template: outputTypeManifestTemplate,
			Input: templatingInput{
				Name:   cfg.ManifestName,
				Prefix: cfg.ManifestsPrefix,
			},
		},
		{
			Template: interfaceManifestTemplate,
			Input: templatingInput{
				Name:   cfg.ManifestName,
				Prefix: cfg.ManifestsPrefix,
			},
		},
	}

	generated, err := generateManifests(cfgs)
	if err != nil {
		return nil, errors.Wrap(err, "while generating manifests")
	}

	result := make(map[string]string, len(generated))

	for _, m := range generated {
		metadata, err := manifest.GetMetadata([]byte(m))
		if err != nil {
			return nil, errors.Wrap(err, "while getting metadata for manifest")
		}

		manifestPath := fmt.Sprintf("%s.%s", metadata.Metadata.Prefix, metadata.Metadata.Name)

		result[manifestPath] = m
	}

	return result, nil
}

const (
	interfaceManifestTemplate = `ocfVersion: 0.0.1
revision: 0.1.0
kind: Interface
metadata:
  prefix: "cap.interface.{{ .Prefix }}"
  name: "{{ .Name }}"
  displayName: "{{ .Name }}"
  description: "{{ .Name }} action for {{ .Prefix }}"
  documentationURL: https://example.com
  supportURL: https://example.com
  iconURL: https://example.com/icon.png
  maintainers:
    - email: dev@example.cop
      name: Example Dev
      url: https://example.com

spec:
  input:
    parameters:
      input-parameters:
        typeRef:
          path: cap.type.{{ .Prefix }}.{{ .Name }}-input
          revision: 0.1.0
    typeInstances: {}

  output:
    typeInstances:
      config:
        typeRef:
          path: cap.type.{{ .Prefix }}.config
          revision: 0.1.0
`
)

const (
	outputTypeManifestTemplate = `ocfVersion: 0.0.1
revision: 0.1.0
kind: Type
metadata:
  name: config
  prefix: "cap.type.{{ .Prefix }}"
  displayName: "{{.Prefix }} config"
  description: "Type representing a {{ .Prefix }} config"
  documentationURL: https://example.com
  supportURL: https://example.com
  maintainers:
    - email: dev@example.com
      name: Example Dev
      url: https://example.com
spec:
  jsonSchema:
    # Put the properties of your Interface output Type in form of a JSON Schema here:
    value: |-
      {
        "$schema": "http://json-schema.org/draft-07/schema",
        "type": "object",
        "required": [],
        "properties": {
          "example": {
            "$id": "#/properties/example",
            "type": "String",
            "description": "Example field"
          }
        }
      }
`
)
