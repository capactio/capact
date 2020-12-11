package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"projectvoltron.dev/voltron/docs/investigation/workflow-rendering/render"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
	"sigs.k8s.io/yaml"
)

var (
	implementationsDir = "docs/investigation/workflow-rendering/implementations"
	implementations    = map[string]*types.Implementation{}
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("missing implementation path argument")
	}

	implementationPath := os.Args[1]

	if err := loadImplementations(); err != nil {
		log.Fatal(err)
	}

	renderer := &render.Renderer{
		Implementations: implementations,
	}

	toRender, ok := implementations[implementationPath]
	if !ok {
		log.Fatalf("implementation %s does not exist", implementationPath)
	}

	data, err := renderer.Render(toRender)
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

	yamlData, err := yaml.Marshal(obj)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(yamlData))
}

func loadImplementations() error {
	fis, err := ioutil.ReadDir(implementationsDir)
	if err != nil {
		return errors.Wrap(err, "while listing implementation dir")
	}

	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}

		filepath := fmt.Sprintf("%s/%s", implementationsDir, fi.Name())
		if err := loadImplementation(filepath); err != nil {
			log.Printf("failed to load implementation %s: %v", filepath, err)
		}
	}

	return nil
}

func loadImplementation(filepath string) error {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return errors.Wrap(err, "while reading file")
	}

	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return errors.Wrap(err, "while converting YAML to JSON")
	}

	impl, err := types.UnmarshalImplementation(jsonData)
	if err != nil {
		return errors.Wrap(err, "while unmarshaling implementation")
	}

	key := fmt.Sprintf("%s.%s", *impl.Metadata.Prefix, impl.Metadata.Name)
	implementations[key] = &impl
	return nil
}
