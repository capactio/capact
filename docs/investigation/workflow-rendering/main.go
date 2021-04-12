package main

import (
	"flag"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"log"
	"capact.io/capact/docs/investigation/workflow-rendering/render"
	"capact.io/capact/pkg/engine/k8s/api/v1alpha1"
)

type renderInput struct {
	Name              string                           `json:"name"`
	ManifestReference v1alpha1.ManifestReference       `json:"manifestReference"`
	Parameters        map[string]interface{}           `json:"parameters"`
	TypeInstances     []*v1alpha1.InputTypeInstance    `json:"typeInstances"`
	Policies          map[string]render.FilterPolicies `json:"policies"`
}

func main() {
	ochDir := flag.String("och-dir", "manifests", "Directory with OCH manifests")
	typeInstanceDir := flag.String("type-instances-dir", "manifests/typeinstances", "Directory with OCH manifests")
	renderInputPath := flag.String("render-input", "", "Filepath to render input")
	flag.Parse()

	if *renderInputPath == "" {
		log.Fatal("missing render-input filepath")
	}

	renderInputData, err := ioutil.ReadFile(*renderInputPath)
	if err != nil {
		log.Fatal(err)
	}

	renderInput := &renderInput{}
	if err := yaml.Unmarshal(renderInputData, renderInput); err != nil {
		log.Fatal(err)
	}

	manifestStore, err := render.NewManifestStore(*ochDir, *typeInstanceDir)
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
		renderInput.Policies,
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
