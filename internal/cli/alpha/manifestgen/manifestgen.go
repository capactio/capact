package manifestgen

import (
	"bytes"
	"fmt"
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

func splitPathToPrefixAndName(path string) (string, string, error) {
	parts := strings.Split(path, ".")
	if len(parts) < 4 {
		return "", "", fmt.Errorf("manifest path must have prefix and name")
	}

	prefix := strings.Join(parts[2:len(parts)-1], ".")
	name := parts[len(parts)-1]

	return prefix, name, nil
}
