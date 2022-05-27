package capact

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/fatih/structs"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/strvals"
	"sigs.k8s.io/yaml"
)

type (
	// InputParameters for Capact Helm charts
	InputParameters struct {
		Version                string `json:"version"`
		IncreaseResourceLimits bool   `json:"-"`
		ActionCRDLocation      string `json:"-"`
		Override               struct {
			CapactStringOverrides      []string
			IngressStringOverrides     []string
			CertManagerStringOverrides []string

			HelmRepo     string      `json:"helmRepoURL"`
			CapactValues Values      `json:"capactValues,omitempty"`
			Neo4jValues  Neo4jValues `json:"neo4jValues,omitempty"`
		} `json:"override"`
	}
	// Neo4jValues for Neo4j Helm chart
	Neo4jValues struct {
		Neo4j struct {
			Core struct {
				Resources Resources `json:"resources,omitempty"`
			} `json:"core,omitempty"`
		} `json:"neo4j,omitempty"`
	}
	// Values for Capact Helm charts
	Values struct {
		Notes struct {
			PrintInsecure bool `json:"printInsecure"`
		} `json:"notes"`
		Engine    Engine    `json:"engine,omitempty"`
		Gateway   Gateway   `json:"gateway,omitempty"`
		HubPublic HubPublic `json:"hub-public,omitempty"`
		HubLocal  HubLocal  `json:"hub-local,omitempty"`
		Global    struct {
			ContainerRegistry struct {
				Tag  string `json:"overrideTag,omitempty"`
				Path string `json:"path,omitempty"`
			} `json:"containerRegistry,omitempty"`
		} `json:"global,omitempty"`
	}
	// ResourcesQuantity resource quantity values
	ResourcesQuantity struct {
		CPU    string `json:"cpu,omitempty"`
		Memory string `json:"memory,omitempty"`
	}
	// Resources values
	Resources struct {
		Limits   ResourcesQuantity `json:"limits,omitempty"`
		Requests ResourcesQuantity `json:"requests,omitempty"`
	}
	// Gateway values
	Gateway struct {
		Resources Resources `json:"resources,omitempty"`
	}
	// HubPublic values
	HubPublic struct {
		Resources Resources `json:"resources,omitempty"`
		Populator Populator `json:"populator,omitempty"`
	}
	// HubLocal values
	HubLocal struct {
		Resources Resources `json:"resources,omitempty"`
	}
	// Engine values
	Engine struct {
		TestSetup struct {
			Enabled bool `json:"enabled,omitempty"`
		} `json:"testSetup,omitempty"`
	}
	//Populator values
	Populator struct {
		Enabled            bool                `json:"enabled,omitempty"`
		ManifestsLocations []ManifestsLocation `json:"manifestsLocations,omitempty"`
	}
	// ManifestsLocation values
	ManifestsLocation struct {
		Branch     string `json:"branch,omitempty"`
		Repository string `json:"repository,omitempty"`
	}
)

// IncreasedGatewayResources returns increased Gateway resources
func IncreasedGatewayResources() Resources {
	return Resources{
		Limits: ResourcesQuantity{
			CPU:    "300m",
			Memory: "128Mi",
		},
		Requests: ResourcesQuantity{
			CPU:    "100m",
			Memory: "48Mi",
		},
	}
}

// IncreasedHubPublicResources returns increased Public Hub resources
func IncreasedHubPublicResources() Resources {
	return Resources{
		Limits: ResourcesQuantity{
			CPU:    "400m",
			Memory: "512Mi",
		},
		Requests: ResourcesQuantity{
			CPU:    "200m",
			Memory: "128Mi",
		},
	}
}

// IncreasedHubLocalResources returns increased Local Hub resources
func IncreasedHubLocalResources() Resources {
	return Resources{
		Limits: ResourcesQuantity{
			CPU:    "400m",
			Memory: "512Mi",
		},
		Requests: ResourcesQuantity{
			CPU:    "200m",
			Memory: "128Mi",
		},
	}
}

// IncreasedNeo4jResources returns increased Neo4j resources
func IncreasedNeo4jResources() Resources {
	return Resources{
		Limits: ResourcesQuantity{
			CPU:    "1",
			Memory: "3072Mi",
		},
		Requests: ResourcesQuantity{
			CPU:    "500m",
			Memory: "1768Mi",
		},
	}
}

// ResolveVersion resolves @latest tag
func (i *InputParameters) ResolveVersion() error {
	if i.Override.HelmRepo == LatestVersionTag {
		i.Override.HelmRepo = HelmRepoLatest
	}

	switch i.Version {
	case LatestVersionTag:
		ver, err := NewHelmHelper().GetLatestVersion(i.Override.HelmRepo, Name)
		if err != nil {
			return err
		}
		i.Version = ver

	case LocalVersionTag:
		i.Override.HelmRepo = LocalChartsPath

		if i.Override.CapactValues.Global.ContainerRegistry.Path == "" {
			i.Override.CapactValues.Global.ContainerRegistry.Path = LocalDockerPath
		}
		if i.Override.CapactValues.Global.ContainerRegistry.Tag == "" {
			i.Override.CapactValues.Global.ContainerRegistry.Tag = LocalDockerTag
		}
	}

	// if not already set via flags, resolve base on our logic
	if i.ActionCRDLocation == "" {
		if err := i.resolveCRDLocationFromVersion(); err != nil {
			return err
		}
	}

	return nil
}

// resolveCRDLocationFromVersion sets the CRD location.
// If version was:
// - local tag, use the relative local CRD path
// - stable release (tag), use tag
// - the latest release from main (tag-commit), use the commit sha
func (i *InputParameters) resolveCRDLocationFromVersion() error {
	if i.Version == LocalVersionTag {
		i.ActionCRDLocation = LocalCRDPath
		return nil
	}

	decoded, err := semver.NewVersion(i.Version)
	if err != nil {
		return errors.Wrap(err, "while parsing SemVer version")
	}
	if decoded.Prerelease() != "" { // version in format {tag-commit}
		i.ActionCRDLocation = fmt.Sprintf(CRDUrlFormat, decoded.Prerelease())
	} else { // version in format {tag}
		ghTag := fmt.Sprintf("v%s", decoded.String())
		i.ActionCRDLocation = fmt.Sprintf(CRDUrlFormat, ghTag)
	}

	return nil
}

// SetCapactValuesFromOverrides fills CapactValues struct with values passed in Override.CapactStringOverrides
func (i *InputParameters) SetCapactValuesFromOverrides() error {
	mapValues := i.Override.CapactValues.AsMap()

	for _, value := range i.Override.CapactStringOverrides {
		if err := strvals.ParseInto(value, mapValues); err != nil {
			return errors.Wrap(err, "failed parsing passed overrides")
		}
	}
	values, err := ValuesFromMap(mapValues)
	if err != nil {
		return errors.Wrap(err, "while converting map to values")
	}
	i.Override.CapactValues = values
	return nil
}

// AsMap converts Values struct into map[string]interface{}
func (i *Values) AsMap() map[string]interface{} {
	s := structs.New(i)
	s.TagName = "json"
	return s.Map()
}

// AsMap converts Values struct into map[string]interface{}
func (n *Neo4jValues) AsMap() map[string]interface{} {
	s := structs.New(n)
	s.TagName = "json"
	return s.Map()
}

// ValuesFromMap returns Values struct converted from map[string]interface{}
func ValuesFromMap(values map[string]interface{}) (Values, error) {
	v := Values{}
	marshaled, err := yaml.Marshal(values)
	if err != nil {
		return v, errors.Wrap(err, "failed to marshal input values")
	}
	err = yaml.Unmarshal(marshaled, &v)
	if err != nil {
		return v, errors.Wrap(err, "failed to unmarshal input values")
	}
	return v, nil
}
