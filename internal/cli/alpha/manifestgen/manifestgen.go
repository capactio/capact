package manifestgen

import (
	"bytes"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"
)

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
	functionMap := template.FuncMap(sprig.FuncMap())
	functionMap["DerefS"] = func(s *string) string { return *s }
	tmpl, err := template.New("manifest").
		Funcs(functionMap).
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
