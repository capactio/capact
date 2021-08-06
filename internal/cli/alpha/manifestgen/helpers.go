package manifestgen

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"
)

// WriteManifestFiles writes the manifests file in files parameters to the provided outputDir.
// Depending on the override parameter is will either override existing manifest files or skip them.
func WriteManifestFiles(outputDir string, files map[string]string, override bool) error {
	for manifestPath, content := range files {
		manifestFilepath := strings.ReplaceAll(manifestPath, ".", string(os.PathSeparator)) + ".yaml"
		outputFilepath := path.Join(outputDir, manifestFilepath)

		if err := os.MkdirAll(path.Dir(outputFilepath), 0750); err != nil {
			return errors.Wrap(err, "while creating directory for generated manifests")
		}

		if _, err := os.Stat(outputFilepath); !override && !os.IsNotExist(err) {
			fmt.Printf("%s Skipped %s as it already exists in %s\n", color.YellowString("-"), manifestPath, outputFilepath)
			continue
		}

		if err := os.WriteFile(outputFilepath, []byte(content), 0600); err != nil {
			return errors.Wrapf(err, "while writing generated manifest %s", manifestPath)
		}

		fmt.Printf("%s Generated %s in %s\n", color.GreenString("✓"), manifestPath, outputFilepath)
	}

	return nil
}
