// +build generate

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

const schemaDir = "../../../../ocf-spec"

func main() {
	fs := http.Dir(schemaDir)

	err := vfsgen.Generate(fs, vfsgen.Options{
		Filename:     "static_schema_gen.go",
		PackageName:  "validate",
		VariableName: "StaticSchema",
	})
	if err != nil {
		log.Fatal(err)
	}
}
