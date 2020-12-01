// +build generate

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/shurcooL/vfsgen"
)

const schemaDir = "../../../../ocf-spec"

// Use the Go built-in functionality after switching to Go 1.16
// More info: https://github.com/golang/go/issues/41191
func main() {
	fs := &HTTPFileSystem{
		fs: http.Dir(schemaDir),
		skipDirs: map[string]struct{}{
			"examples": {},
		},
	}

	err := vfsgen.Generate(fs, vfsgen.Options{
		Filename:     "static_schema_gen.go",
		PackageName:  "validate",
		VariableName: "StaticSchema",
	})
	if err != nil {
		log.Fatal(err)
	}
}

// HTTPFileSystem implements http.FileSystem interface and provides functionality to skip a given directories.
type HTTPFileSystem struct {
	fs       http.FileSystem
	skipDirs map[string]struct{}
}

func (fs HTTPFileSystem) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return dirsIgnorant{f, fs.skipDirs}, nil
}

type dirsIgnorant struct {
	http.File
	skipDirs map[string]struct{}
}

func (f dirsIgnorant) Stat() (os.FileInfo, error) {
	fi, err := f.File.Stat()
	if err != nil {
		return nil, err
	}
	_, skip := f.skipDirs[fi.Name()]
	if fi.IsDir() && skip {
		return nil, fmt.Errorf("directory %q should be ignored", fi.Name())
	}
	return fi, nil
}
