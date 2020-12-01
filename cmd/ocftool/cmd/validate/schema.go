package validate

import (
	"net/http"
	"os"

	"github.com/spf13/pflag"
)

//go:generate go run embed_schema.go

// SchemaProvider provides functionality to read embedded static schema or schema from local file system.
type SchemaProvider struct {
	localSchemaRootDir string
}

// FileSystem returns file system implementation and root dir for schema directory.
func (s *SchemaProvider) FileSystem() (http.FileSystem, string) {
	if len(s.localSchemaRootDir) != 0 {
		return &LocalSchema{}, s.localSchemaRootDir
	}

	return StaticSchema, "."
}

// RegisterSchemaFlags registers schema related flags
func (s *SchemaProvider) RegisterSchemaFlags(flagSet *pflag.FlagSet) {
	flagSet.StringVarP(&s.localSchemaRootDir, "schemas", "s", "", "Path to the local directory with OCF JSONSchemas. If not provided, built-in JSONSchemas are used.")
}

// LocalSchema fulfils the http.FileSystem interface and provides functionality to opens files from local file system.
type LocalSchema struct{}

// Open opens the named file for reading.
func (*LocalSchema) Open(name string) (http.File, error) {
	return os.Open(name)
}
