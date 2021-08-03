package content

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
)

func writeManifestFiles(files map[string]string) error {
	for manifestPath, content := range files {
		manifestFilepath := path.Join("generated", strings.ReplaceAll(manifestPath, ".", string(os.PathSeparator))+".yaml")

		if err := os.MkdirAll(path.Dir(manifestFilepath), 0750); err != nil {
			return errors.Wrap(err, "while creating directory for generated manifests")
		}

		if err := os.WriteFile(manifestFilepath, []byte(content), 0600); err != nil {
			return errors.Wrapf(err, "while writing generated manifest %s", manifestPath)
		}

		fmt.Printf("Generated %s in %s\n", manifestPath, manifestFilepath)
	}

	return nil
}
