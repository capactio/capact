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
		manifestFilepath := strings.ReplaceAll(manifestPath, ".", string(os.PathSeparator)) + ".yaml"
		outputFilepath := path.Join(manifestOutputDirectory, manifestFilepath)

		if err := os.MkdirAll(path.Dir(outputFilepath), 0750); err != nil {
			return errors.Wrap(err, "while creating directory for generated manifests")
		}

		if err := os.WriteFile(outputFilepath, []byte(content), 0600); err != nil {
			return errors.Wrapf(err, "while writing generated manifest %s", manifestPath)
		}

		fmt.Printf("Generated %s in %s\n", manifestPath, outputFilepath)
	}

	return nil
}
