package argo

import (
	"github.com/ghodss/yaml"
	"io/ioutil"
	"log"
	"projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
	"projectvoltron.dev/voltron/pkg/och/client/fake"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
	"testing"
)

func Benchmark(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Render()
	}
}

type renderInput struct {
	Name              string                       `json:"name"`
	ManifestReference types.TypeRef                `json:"manifestReference"`
	Parameters        map[string]interface{}       `json:"parameters"`
	TypeInstances     []v1alpha1.InputTypeInstance `json:"typeInstances"`
}

func Render() {
	manifestStore, err := fake.NewFromLocal("testdata/och")
	if err != nil {
		log.Fatal(err)
	}
	renderInputData, err := ioutil.ReadFile("testdata/1-postgres.yml")
	if err != nil {
		log.Fatal(err)
	}

	renderInput := &renderInput{}
	if err := yaml.Unmarshal(renderInputData, renderInput); err != nil {
		log.Fatal(err)
	}
	renderer := NewRenderer(manifestStore, WithPlainTextUserInput(renderInput.Parameters), WithTypeInstances(renderInput.TypeInstances))

	_, err = renderer.Render(
		renderInput.ManifestReference,
	)
	if err != nil {
		log.Fatal(err)
	}
}
