package schema

import (
	"net/http"
	"os"
	"path/filepath"
)

//go:generate go run embed_schema.go

// Provider provides functionality to read embedded static schema or schema from local file system.
type Provider struct {
	localSchemaRootDir string
}

// NewProvider instantiates new Provider.
func NewProvider(localSchemaRootDir string) *Provider {
	return &Provider{localSchemaRootDir: localSchemaRootDir}
}

// FileSystem returns file system implementation and root dir for schema directory.
func (s *Provider) FileSystem() (http.FileSystem, string) {
	if len(s.localSchemaRootDir) != 0 {
		return &LocalFileSystem{}, s.localSchemaRootDir
	}

	return Static, "."
}

// LocalFileSystem fulfils the http.FileSystem interface and provides functionality to open files from local file system.
type LocalFileSystem struct{}

// Open opens the named file for reading.
func (*LocalFileSystem) Open(name string) (http.File, error) {
	return os.Open(filepath.Clean(name))
}
