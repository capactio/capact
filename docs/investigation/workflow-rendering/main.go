package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/mitchellh/mapstructure"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"projectvoltron.dev/voltron/docs/investigation/workflow-rendering/render"
	"projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
)

var (
	implementationsDir = "docs/investigation/workflow-rendering/manifests"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("missing implementation path argument")
	}

	implementationPath := os.Args[1]

	manifestStore, err := render.NewManifestStore(implementationsDir)
	if err != nil {
		log.Fatal(err)
	}

	renderer := &render.Renderer{
		ManifestStore: manifestStore,
	}

	toRender := manifestStore.GetImplementation(v1alpha1.ManifestReference{
		Path: v1alpha1.NodePath(implementationPath),
	})
	if toRender == nil {
		log.Fatalf("implementation %s does not exist", implementationPath)
	}

	data, err := renderer.Render(
		toRender,
		map[string]interface{}{
			"superuser": map[string]interface{}{
				"username": "postgres",
				"password": "s3cr3t",
			},
			"defaultDBName": "test",
		},
		[]*v1alpha1.InputTypeInstance{
			//{
			//	Name: "postgresql",
			//	ID:   "461a1c83-6054-43dd-8a4c-49acde791699",
			//},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	obj := &unstructured.Unstructured{}

	workflowName := strings.Replace(implementationPath, ".", "-", -1)

	obj.SetKind("Workflow")
	obj.SetAPIVersion("argoproj.io/v1alpha1")
	obj.SetName(workflowName)

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
