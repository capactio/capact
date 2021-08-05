package manifestgen

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

// Config stores the generic input parameters for content generation
type Config struct {
	ManifestPath     string
	ManifestRevision string
}

type templatingConfig struct {
	Template string
	Input    interface{}
}

type templatingInput struct {
	Name      string
	Prefix    string
	Revision  string
	Variables []inputVariable
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

var tmplFuncs = template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
}

func generateManifests(cfgs []*templatingConfig) ([]string, error) {
	manifests := make([]string, 0, len(cfgs))

	for _, cfg := range cfgs {
		manifest, err := generateManifest(cfg)
		if err != nil {
			return nil, errors.Wrapf(err, "while generating manifest: %v", err)
		}

		manifests = append(manifests, manifest)
	}

	return manifests, nil
}

func generateManifest(cfg *templatingConfig) (string, error) {
	tmpl, err := template.New("manifest").
		Funcs(tmplFuncs).
		Parse(cfg.Template)
	if err != nil {
		return "", errors.Wrap(err, "while creating new template")
	}

	var manifest bytes.Buffer
	if err := tmpl.Execute(&manifest, cfg.Input); err != nil {
		return "", errors.Wrap(err, "while executing Go template")
	}

	return manifest.String(), nil
}

func splitPathToPrefixAndName(path string) (string, string) {
	parts := strings.Split(path, ".")
	prefix := strings.Join(parts[2:len(parts)-1], ".")
	name := parts[len(parts)-1]

	return prefix, name
}

const (
	typeManifestTemplate = `ocfVersion: 0.0.1
revision: 0.1.0
kind: Type
metadata:
  name: {{ .Name }}-input
  prefix: "cap.type.{{ .Prefix }}"
  displayName: "Input for {{ .Prefix }}.{{ .Name }}"
  description: Input for the "{{ .Prefix }}.{{ .Name }} Action"
  documentationURL: https://example.com
  supportURL: https://example.com
  maintainers:
    - email: dev@example.com
      name: Example Dev
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
            "description": "{{ $variable.Description }}"{{if $variable.Default }},
            "default": "{{ $variable.Default }}"{{ end }}
          }{{ if ne $index (add $length -1) }},{{ end }}
          {{ end }}
        }
      }
`
)
