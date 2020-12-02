package schema

import (
	"net/http"
	"os"

	"github.com/spf13/pflag"
)

//go:generate go run embed_schema.go

// Provider provides functionality to read embedded static schema or schema from local file system.
type Provider struct {
	localSchemaRootDir string
}

// FileSystem returns file system implementation and root dir for schema directory.
func (s *Provider) FileSystem() (http.FileSystem, string) {
	if len(s.localSchemaRootDir) != 0 {
		return &LocalFileSystem{}, s.localSchemaRootDir
	}

	return Static, "."
}

// RegisterSchemaFlags registers schema related flags
func (s *Provider) RegisterSchemaFlags(flagSet *pflag.FlagSet) {
	flagSet.StringVarP(&s.localSchemaRootDir, "schemas", "s", "", "Path to the local directory with OCF JSONSchemas. If not provided, built-in JSONSchemas are used.")
}

// LocalFileSystem fulfils the http.FileSystem interface and provides functionality to open files from local file system.
type LocalFileSystem struct{}

// Open opens the named file for reading.
func (*LocalFileSystem) Open(name string) (http.File, error) {
	return os.Open(name)
}
