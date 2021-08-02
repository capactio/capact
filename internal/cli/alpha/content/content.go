package content

import (
	"bytes"
	"text/template"

	"github.com/pkg/errors"
)

// Config stores the generic input parameters for content generation
type Config struct {
	ModulePath      string
	ManifestName    string
	ManifestsPrefix string
	InterfacePath   string
}

type inputVariable struct {
	Name        string
	Type        string
	Description string
	Default     interface{}
}

type outputVariable struct {
	Name string
}

type templatingInput struct {
	Name          string
	Prefix        string
	InterfacePath string
	Variables     []inputVariable
	Outputs       []outputVariable
}

var tmplFuncs = template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
}

func generateManifests(input *templatingInput, manifestTemplates map[string]string) (map[string]string, error) {
	manifests := make(map[string]string)

	for manifestPath, tmpl := range manifestTemplates {
		manifest, err := generateManifest(input, tmpl)
		if err != nil {
			return nil, errors.Wrapf(err, "while generating manifest %s", manifestPath)
		}

		manifests[manifestPath] = manifest
	}

	return manifests, nil
}

func generateManifest(input *templatingInput, templateString string) (string, error) {
	typeTemplate, err := template.New("manifest").
		Funcs(tmplFuncs).
		Parse(templateString)
	if err != nil {
		return "", errors.Wrap(err, "while creating new template")
	}

	var typeManifest bytes.Buffer
	if err := typeTemplate.Execute(&typeManifest, input); err != nil {
		return "", errors.Wrap(err, "while executing Go template")
	}

	return typeManifest.String(), nil
}

const (
	additonalInputTypeTemplate = `
ocfVersion: 0.0.1
revision: 0.1.0
kind: Type
metadata:
  name: {{ .Name }}-input
  prefix: "cap.type.{{ .Prefix }}"
  displayName: "Additional input for {{ .Prefix }}.{{ .Name }}"
  description: Additional input for the "{{ .Prefix }}.{{ .Name }} Action"
  documentationURL: https://example.com
  supportURL: https://example.com
  maintainers:
    - email: dev@example.com
      name: Example User
      url: https://example.com
spec:
  jsonSchema:
    value: |-
      {
        "$schema": "http://json-schema.org/draft-07/schema",
        "type": "object",
        "required": [],
        "properties": {
          {{ $length := len .Variables -}}
          {{ range $index, $variable := .Variables -}}
          "{{ $variable.Name }}": {
            "$id": "#/properties/{{ $variable.Name }}",
            "type": "{{ $variable.Type }}",
            "description": "{{ $variable.Description }}",
            "default": "{{ $variable.Default }}"
          }{{if ne $index (add $length -1) }},{{end}}
          {{ end }}
        }
      }
`
)
