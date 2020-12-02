// +build generate

package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/shurcooL/httpfs/filter"
	"github.com/shurcooL/vfsgen"
)

const schemaDir = "../../../ocf-spec"

// TODO: Use the Go built-in functionality after switching to Go 1.16
// More info: https://github.com/golang/go/issues/41191
func main() {
	fs := filter.Skip(
		http.Dir(schemaDir),
		func(path string, fi os.FileInfo) bool {
			return fi.Name() == "README.md" || (fi.IsDir() && fi.Name() == "examples")
		})

	fs = zeroTimeFileSystem{fs}

	err := vfsgen.Generate(fs, vfsgen.Options{
		Filename:     "static_schema_gen.go",
		PackageName:  "schema",
		VariableName: "Static",
	})
	if err != nil {
		log.Fatal(err)
	}
}

// zeroTimeFileSystem implements http.FileSystem interface and provides functionality to skip modTime.
// vfsgen can't generate deterministic content when executed from different systems (because file timestamps aren't stable).
// Fortunately, we do not need that as we run validation for generated files and each modification in content will be detected.
// More info in issue: https://github.com/shurcooL/vfsgen/issues/26
type zeroTimeFileSystem struct {
	fs http.FileSystem
}

func (fs zeroTimeFileSystem) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return &zeroTimeStat{f}, nil
}

type zeroTimeStat struct {
	http.File
}

func (f *zeroTimeStat) Stat() (os.FileInfo, error) {
	info, err := f.File.Stat()
	if err != nil {
		return nil, err
	}
	return &zeroModTime{info}, nil
}

type zeroModTime struct {
	os.FileInfo
}

func (z *zeroModTime) ModTime() time.Time {
	return time.Time{}
}
