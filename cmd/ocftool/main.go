package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"github.com/xeipuuv/gojsonschema"
)

const (
	appName = "ocftool"
)

func schemaFileForKind(kind string) (string, error) {
	schemaFilepath, ok := schemaPaths[kind]
	if !ok {
		return "", fmt.Errorf("cannot find schema file for kind %s", kind)
	}
	return schemaFilepath, nil
}

var (
	commonSchemaPaths []string = []string{
		"common/json-schema-type.json",
		"common/metadata-tags.json",
		"common/metadata.json",
	}

	schemaPaths = map[string]string{
		"Implementation": "implementation.json",
		"InterfaceGroup": "interface-group.json",
		"Interface":      "interface.json",
		"RepoMetadata":   "repo-metadata.json",
		"Tag":            "tag.json",
		"Type":           "type.json",
		"TypeInstance":   "type-instance.json",
		"Vendor":         "vendor.json",
	}

	manifestFilepath string

	rootCmd = &cobra.Command{
		Use: appName,

		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				panic(err)
			}
		},
	}

	validateCmd = &cobra.Command{
		Use:  "validate",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			for _, filepath := range args {
				sl := gojsonschema.NewSchemaLoader()

				for _, path := range commonSchemaPaths {
					jsonLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://ocf-spec/0.0.1/schema/%s", path))
					err := sl.AddSchemas(jsonLoader)
					if err != nil {
						panic(err)
					}
				}

				manifestBytes, err := ioutil.ReadFile(filepath)
				if err != nil {
					// TODO error handling
					panic(err)
				}

				manifestKind := &ManifestKind{}

				err = yaml.Unmarshal(manifestBytes, manifestKind)
				if err != nil {
					panic(err)
				}

				schemaRootFile, err := schemaFileForKind(manifestKind.Kind)
				if err != nil {
					panic(err)
				}
				schemaRootLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://ocf-spec/0.0.1/schema/%s", schemaRootFile))

				schema, err := sl.Compile(schemaRootLoader)
				if err != nil {
					panic(err)
				}

				manifestJsonBytes, err := yaml.YAMLToJSON(manifestBytes)
				if err != nil {
					panic(err)
				}

				manifestLoader := gojsonschema.NewBytesLoader(manifestJsonBytes)

				res, err := schema.Validate(manifestLoader)
				if err != nil {
					panic(err)
				}

				if !res.Valid() {
					for _, err := range res.Errors() {
						fmt.Println(err)
					}
				}
			}
		},
	}
)

func main() {
	cobra.OnInitialize()
	rootCmd.AddCommand(validateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
