package manifestgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/jsonschema"
	"github.com/iancoleman/orderedmap"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

// GenerateHelmManifests generates manifest files for a Helm module based Implementation
func GenerateHelmManifests(cfg *HelmConfig) (map[string]string, error) {
	helmChart, err := loadHelmChart(cfg)
	if err != nil {
		return nil, err
	}

	cfgs := make([]*templatingConfig, 0, 2)

	inputTypeCfg, err := getHelmInputTypeTemplatingConfig(cfg, helmChart)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Helm templating input")
	}
	cfgs = append(cfgs, inputTypeCfg)

	implCfg, err := getHelmImplementationTemplatingConfig(cfg, helmChart)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Helm templating input")
	}
	cfgs = append(cfgs, implCfg)

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
	cpo.RepoURL = cfg.ChartRepoURL
	cpo.Version = cfg.ChartVersion

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

func getHelmInputTypeTemplatingConfig(cfg *HelmConfig, helmChart *chart.Chart) (*templatingConfig, error) {
	prefix, name, err := splitPathToPrefixAndName(cfg.ManifestPath)
	if err != nil {
		return nil, errors.Wrap(err, "while getting prefix and path for manifests")
	}

	jsonSchema, err := getJSONSchemaForValues(helmChart)
	if err != nil {
		return nil, errors.Wrap(err, "while getting JSON schema for Helm chart values")
	}

	input := &typeTemplatingInput{
		templatingInput: templatingInput{
			Name:     name,
			Prefix:   prefix,
			Revision: cfg.ManifestRevision,
		},
		JSONSchema: string(jsonSchema),
	}

	return &templatingConfig{
		Template: typeManifestTemplate,
		Input:    input,
	}, nil
}

func getHelmImplementationTemplatingConfig(cfg *HelmConfig, helmChart *chart.Chart) (*templatingConfig, error) {
	var (
		helmValues        = make(map[string]interface{})
		interfacePath     = cfg.InterfacePathWithRevision
		interfaceRevision = "0.1.0"
	)

	prefix, name, err := splitPathToPrefixAndName(cfg.ManifestPath)
	if err != nil {
		return nil, errors.Wrap(err, "while getting prefix and path for manifests")
	}

	pathSlice := strings.SplitN(cfg.InterfacePathWithRevision, ":", 2)
	if len(pathSlice) == 2 {
		interfacePath = pathSlice[0]
		interfaceRevision = pathSlice[1]
	}

	if err := deepCopy(&helmValues, helmChart.Values); err != nil {
		return nil, errors.Wrap(err, "while deep copying Helm values")
	}

	if err := injectJinjaTemplatingToHelmValues(helmValues, []string{}); err != nil {
		return nil, errors.Wrap(err, "while setting values for Helm input values")
	}

	var helmValuesBuf bytes.Buffer
	enc := yaml.NewEncoder(&helmValuesBuf)
	enc.SetIndent(2)

	if err := enc.Encode(helmValues); err != nil {
		return nil, errors.Wrap(err, "while marshaling Helm runner values")
	}

	input := &helmImplementationTemplatingInput{
		templatingInput: templatingInput{
			Name:     name,
			Prefix:   prefix,
			Revision: cfg.ManifestRevision,
		},
		InterfacePath:     interfacePath,
		InterfaceRevision: interfaceRevision,
		HelmChartName:     helmChart.Name(),
		HelmChartVersion:  helmChart.Metadata.Version,
		HelmRepoURL:       cfg.ChartRepoURL,
		ValuesYAML:        helmValuesBuf.String(),
	}

	return &templatingConfig{
		Template: helmImplementationManifestTemplate,
		Input:    input,
	}, nil
}

func injectJinjaTemplatingToHelmValues(values map[string]interface{}, parentKeyPath []string) error {
	for key, v := range values {
		keyPath := append(parentKeyPath, key)
		keyPathString := buildValueKeyPathForJinja(keyPath)

		switch value := v.(type) {
		case map[string]interface{}:
			if err := injectJinjaTemplatingToHelmValues(value, keyPath); err != nil {
				return err
			}
		case string:
			values[key] = fmt.Sprintf(`<@ additionalinput.%s | default("%v") | tojson @>`, keyPathString, value)
		case bool:
			values[key] = fmt.Sprintf(`<@ additionalinput.%s | default(%v) | tojson @>`, keyPathString, value)
		case float64:
			values[key] = fmt.Sprintf(`<@ additionalinput.%s | default(%v) @>`, keyPathString, value)
		case []interface{}:
			if value == nil {
				values[key] = fmt.Sprintf(`<@ additionalinput.%s | default(None | tojson) @>`, keyPathString)
				break
			}

			sliceBytes, err := json.Marshal(value)
			if err != nil {
				return errors.Wrapf(err, "while marshaling slice %v", value)
			}

			values[key] = fmt.Sprintf(`<@ additionalinput.%s | default(%v) @>`, keyPathString, string(sliceBytes))
		default:
			values[key] = fmt.Sprintf(`<@ additionalinput.%s | default(%v) | tojson @>`, keyPathString, value)
		}
	}

	return nil
}

func buildValueKeyPathForJinja(keys []string) string {
	if len(keys) == 0 {
		return ""
	}

	acc := keys[0]

	for _, key := range keys[1:] {
		if strings.ContainsRune(key, '.') {
			newAcc := acc + fmt.Sprintf(`["%s"]`, key)
			fmt.Fprintf(os.Stderr, "WARNING: Helm chart values key contains a dot, so we fallback to square brackets in \"%s\". You have to manually check, if %s is not undefined using \"<%% if %s -%%>...<%%- endif %%>\"!\n", newAcc, acc, acc)
			acc = newAcc
		} else {
			acc += fmt.Sprintf(".%s", key)
		}
	}

	return acc
}

func getJSONSchemaForValues(helmChart *chart.Chart) ([]byte, error) {
	if helmChart.Schema != nil {
		return helmChart.Schema, nil
	}

	jsonSchema := generateJSONSchemaForValue(helmChart.Values, []string{"#"})

	schemaBytes, err := json.MarshalIndent(jsonSchema, "", "  ")
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling jsonSchema")
	}

	return schemaBytes, nil
}

func generateJSONSchemaForValue(value interface{}, parentKeyPath []string) *jsonschema.Type {
	ID := strings.Join(parentKeyPath, "/properties/")

	schema := &jsonschema.Type{
		Title: "",
		Extras: map[string]interface{}{
			"$id": ID,
		},
	}

	if len(parentKeyPath) > 0 {
		schema.Title = strcase.ToCamel(parentKeyPath[len(parentKeyPath)-1])
	}

	switch v := value.(type) {
	case string:
		schema.Type = "string"
		schema.Default = v

	case float64:
		schema.Type = "number"
		schema.Default = v

	case bool:
		schema.Type = "boolean"
		schema.Default = v

	case []interface{}:
		schema.Type = "array"
		schema.Default = v

	case map[string]interface{}:
		schema.Type = "object"
		schema.Properties = orderedmap.New()

		for k, val := range v {
			propSchema := generateJSONSchemaForValue(val, append(parentKeyPath, k))
			schema.Properties.Set(k, propSchema)
		}

		schema.Properties.Sort(func(a, b *orderedmap.Pair) bool {
			return a.Key() < b.Key()
		})
	}

	return schema
}
