package render

import (
	"github.com/ghodss/yaml"
	"io/ioutil"
	"log"
	"projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
	"testing"
)


func Benchmark(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Render()
	}
}

type renderInput struct {
	Name              string                           `json:"name"`
	ManifestReference v1alpha1.ManifestReference       `json:"manifestReference"`
	Parameters        map[string]interface{}           `json:"parameters"`
	TypeInstances     []*v1alpha1.InputTypeInstance    `json:"typeInstances"`
	Policies          map[string]FilterPolicies `json:"policies"`
}

func Render() {
	manifestStore, err := NewManifestStore("../../../../och-content/", "../manifests")
	if err != nil {
		log.Fatal(err)
	}
	renderInputData, err := ioutil.ReadFile("../inputs/1-postgres.yml")
	if err != nil {
		log.Fatal(err)
	}

	renderInput := &renderInput{}
	if err := yaml.Unmarshal(renderInputData, renderInput); err != nil {
		log.Fatal(err)
	}
	renderer := &Renderer{
		ManifestStore: manifestStore,
	}

	_, err = renderer.Render(
		renderInput.ManifestReference,
		renderInput.Parameters,
		renderInput.TypeInstances,
		renderInput.Policies,
	)
	if err != nil {
		log.Fatal(err)
	}
}
