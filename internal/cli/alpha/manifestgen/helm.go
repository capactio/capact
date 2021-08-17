package manifestgen

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"sigs.k8s.io/yaml"
)

// HelmConfig stores input parameters for Helm based content generation.
type HelmConfig struct {
	Config

	ChartName string
	RepoURL   string
	Version   string

	InterfacePathWithRevision string
}

type helmTemplatingInput struct {
	templatingInput

	InterfacePath     string
	InterfaceRevision string

	HelmChartName    string
	HelmChartVersion string
	HelmRepoURL      string

	Values interface{}
}

// GenerateHelmManifests generates manifest files for a Helm module based Implementation
func GenerateHelmManifests(cfg *HelmConfig) (map[string]string, error) {
	input, err := getHelmTemplatingInput(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Helm templating input")
	}

	cfgs := []*templatingConfig{
		{
			Template: typeManifestTemplate,
			Input:    input,
		},
		{
			Template: helmImplementationManifestTemplate,
			Input:    input,
		},
	}

	generated, err := generateManifests(cfgs)
	if err != nil {
		return nil, errors.Wrap(err, "while generating manifests")
	}

	result := make(map[string]string, len(generated))

	for _, m := range generated {
		metadata, err := unmarshalMetadata([]byte(m))
		if err != nil {
			return nil, errors.Wrap(err, "while getting metadata for manifest")
		}

		manifestPath := fmt.Sprintf("%s.%s", metadata.Metadata.Prefix, metadata.Metadata.Name)

		result[manifestPath] = m
	}

	return result, nil
}

func loadHelmChart(cfg *HelmConfig) (*chart.Chart, error) {
	cpo := action.ChartPathOptions{}
	cpo.RepoURL = cfg.RepoURL
	cpo.Version = cfg.Version

	chartLocation, err := cpo.LocateChart(cfg.ChartName, &cli.EnvSettings{
		RepositoryCache: "/tmp/helm",
	})
	if err != nil {
		return nil, errors.Wrap(err, "while locating Helm chart")
	}

	chart, err := loader.Load(chartLocation)
	if err != nil {
		return nil, errors.Wrap(err, "while loading Helm chart")
	}

	return chart, nil
}

func getHelmTemplatingInput(cfg *HelmConfig) (*helmTemplatingInput, error) {
	helmChart, err := loadHelmChart(cfg)
	if err != nil {
		return nil, err
	}

	var (
		interfacePath     = cfg.InterfacePathWithRevision
		interfaceRevision = "0.1.0"
		helmValues        = make(map[string]interface{})
	)

	pathSlice := strings.Split(cfg.InterfacePathWithRevision, ":")
	if len(pathSlice) == 2 {
		interfacePath = pathSlice[0]
		interfaceRevision = pathSlice[1]
	}

	prefix, name, err := splitPathToPrefixAndName(cfg.ManifestPath)
	if err != nil {
		return nil, errors.Wrap(err, "while getting prefix and path for manifests")
	}

	if err := deepCopy(&helmValues, helmChart.Values); err != nil {
		return nil, errors.Wrap(err, "while deep copying Helm values")
	}

	if err := deepSetHelmValues(helmChart.Values, []string{}); err != nil {
		return nil, errors.Wrap(err, "while setting values for Helm input values")
	}

	valuesBytes, err := yaml.Marshal(helmChart.Values)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling values YAML")
	}

	return &helmTemplatingInput{
		templatingInput: templatingInput{
			Name:     name,
			Prefix:   prefix,
			Revision: cfg.ManifestRevision,
		},
		InterfacePath:     interfacePath,
		InterfaceRevision: interfaceRevision,
		HelmChartName:     helmChart.Name(),
		HelmChartVersion:  helmChart.Metadata.Version,
		HelmRepoURL:       cfg.RepoURL,
		Values:            string(valuesBytes),
	}, err
}

func deepSetHelmValues(values map[string]interface{}, parentKeyPath []string) error {
	for key, v := range values {
		keyPath := append(parentKeyPath, key)
		keyPathString := strings.Join(keyPath, ".")

		switch value := v.(type) {
		case map[string]interface{}:
			if err := deepSetHelmValues(value, keyPath); err != nil {
				return err
			}
		case string:
			if value == "" {
				// Needed, so empty string will not be interpreted as null in evaluated YAML.
				// TODO: Unfortunately, it does not cover the scenario, where user provides empty string to a parameter.
				value = "''"
			}

			values[key] = fmt.Sprintf(`<@ additionalInput.%s | default("%v") @>`, keyPathString, value)

		case bool:
			values[key] = fmt.Sprintf(`<@ additionalInput.%s | default(%v) | tojson @>`, keyPathString, value)
		case float64:
			values[key] = fmt.Sprintf(`<@ additionalInput.%s | default(%v) @>`, keyPathString, value)
		case []interface{}:
			sliceBytes, err := json.Marshal(value)
			if err != nil {
				return errors.Wrapf(err, "while marshaling slice %v", value)
			}

			values[key] = fmt.Sprintf(`<@ additionalInput.%s | default(%v) @>`, keyPathString, string(sliceBytes))
		default:
			values[key] = fmt.Sprintf(`<@ additionalInput.%s | default(%s) @>`, keyPathString, value)
		}
	}

	return nil
}

//func generateValueJSONSchema(value interface{}, parentKeyPath []string) *jsonschema.Type {
//	schema := &jsonschema.Type{
//		Title: "",
//	}
//
//	if len(parentKeyPath) > 0 {
//		schema.Title = parentKeyPath[len(parentKeyPath)-1]
//	}
//
//	switch v := value.(type) {
//	case string:
//		schema.Type = "string"
//		schema.Default = v
//
//	case map[string]interface{}:
//		schema.Properties = orderedmap.New()
//
//		for k, val := range v {
//			propSchema := generateValueJSONSchema(val, append(parentKeyPath, k))
//			schema.Properties.Set(k, propSchema)
//		}
//	}
//
//	return schema
//}

func deepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}
