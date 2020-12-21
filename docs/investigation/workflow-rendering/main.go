package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/ghodss/yaml"
	"github.com/mitchellh/mapstructure"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"projectvoltron.dev/voltron/docs/investigation/workflow-rendering/render"
	"projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
)

var (
	implementationsDir = "manifests"
)

type renderInput struct {
	Name              string                        `json:"name"`
	ManifestReference v1alpha1.ManifestReference    `json:"manifestReference"`
	Parameters        map[string]interface{}        `json:"parameters"`
	TypeInstances     []*v1alpha1.InputTypeInstance `json:"typeInstances"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("missing implementation path argument")
	}

	inputPath := os.Args[1]

	renderInputData, err := ioutil.ReadFile(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	renderInput := &renderInput{}
	if err := yaml.Unmarshal(renderInputData, renderInput); err != nil {
		log.Fatal(err)
	}

	manifestStore, err := render.NewManifestStore(implementationsDir)
	if err != nil {
		log.Fatal(err)
	}

	renderer := &render.Renderer{
		ManifestStore: manifestStore,
	}

	data, err := renderer.Render(
		renderInput.ManifestReference,
		renderInput.Parameters,
		renderInput.TypeInstances,
	)
	if err != nil {
		log.Fatal(err)
	}

	obj := &unstructured.Unstructured{}

	obj.SetKind("Workflow")
	obj.SetAPIVersion("argoproj.io/v1alpha1")
	obj.SetName(renderInput.Name)

	if err := mapstructure.Decode(map[string]interface{}{
		"spec": data,
	}, &obj.Object); err != nil {
		log.Fatal(err)
	}

	yamlData, err := yaml.Marshal(obj.Object)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(yamlData))
}
