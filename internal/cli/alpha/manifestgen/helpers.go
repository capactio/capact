package manifestgen

import (
	"bytes"
	"encoding/gob"
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
		manifestFilepath := strings.ReplaceAll(strings.TrimPrefix(manifestPath, "cap."), ".", string(os.PathSeparator)) + ".yaml"
		outputFilepath := path.Join(outputDir, manifestFilepath)

		if err := os.MkdirAll(path.Dir(outputFilepath), 0750); err != nil {
			return errors.Wrap(err, "while creating directory for generated manifests")
		}

		if _, err := os.Stat(outputFilepath); !override && !os.IsNotExist(err) {
			fmt.Printf("%s Skipped %q as it already exists in %q\n", color.YellowString("-"), manifestPath, outputFilepath)
			continue
		}

		if err := os.WriteFile(outputFilepath, []byte(content), 0600); err != nil {
			return errors.Wrapf(err, "while writing generated manifest %q", manifestPath)
		}

		fmt.Printf("%s Generated %q in %q\n", color.GreenString("✓"), manifestPath, outputFilepath)
	}

	return nil
}

func deepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func splitPathToPrefixAndName(path string) (string, string, error) {
	parts := strings.Split(path, ".")
	if len(parts) < 3 {
		return "", "", fmt.Errorf("manifest path must have prefix and name")
	}

	prefix := strings.Join(parts[2:len(parts)-1], ".")
	name := parts[len(parts)-1]

	return prefix, name, nil
}
