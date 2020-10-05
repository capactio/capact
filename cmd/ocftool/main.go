package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/xeipuuv/gojsonschema"
	"log"
)

var (
	jsonSchema  = pflag.String("schema", "", "Path to the JSON Schema")
	jsonObjects = pflag.String("objects", "", "Path to the JSON object for validation against JSON schema.")
)

func main() {
	pflag.Parse()

	//if *jsonSchema == "" {
	//
	//}
	//schemaLoader := gojsonschema.NewReferenceLoader("file://pkg/apis/0.0.1/schema/type.json")
	//documentLoader := gojsonschema.NewReferenceLoader("file://pkg/apis/0.0.1/examples/type.json")

	schemaLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", *jsonSchema))
	documentLoader := gojsonschema.NewReferenceLoader(*jsonObjects)

	schema, err := gojsonschema.NewSchema(schemaLoader)
	exitOnError(err, "while creating schema validator")

	result, err := schema.Validate(documentLoader)
	exitOnError(err, "while validating object against JSON Schema")

	if result.Valid() {
		fmt.Printf("The document is valid\n")
	} else {
		fmt.Printf("The document is not valid. see errors:\n")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc.String())
		}
	}
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
