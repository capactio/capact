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
func WriteManifestFiles(outputDir string, files ManifestCollection, override bool) error {
	for manifestPath, content := range files {
		manifestFilepath := createFilePathFromManifestPath(manifestPath)
		outputFilepath := path.Join(outputDir, manifestFilepath)

		if err := os.MkdirAll(path.Dir(outputFilepath), 0750); err != nil {
			return errors.Wrap(err, "while creating directory for generated manifests")
		}

		if _, err := os.Stat(outputFilepath); !override && !os.IsNotExist(err) {
			fmt.Printf("%s Skipped %q as it already exists in %q\n", color.YellowString("-"), manifestPath, outputFilepath)
			continue
		}

		if err := os.WriteFile(outputFilepath, content, 0600); err != nil {
			return errors.Wrapf(err, "while writing generated manifest %q", manifestPath)
		}

		fmt.Printf("%s Generated %q in %q\n", color.GreenString("✓"), manifestPath, outputFilepath)
	}

	return nil
}

//DoesAnyManifestAlreadyExistInDir if any of provided manifests exists in dir.
func DoesAnyManifestAlreadyExistInDir(files ManifestCollection, dir string) bool {
	for manifestPath := range files {
		manifestFilepath := createFilePathFromManifestPath(manifestPath)
		outputFilepath := path.Join(dir, manifestFilepath)
		if _, err := os.Stat(outputFilepath); !os.IsNotExist(err) {
			return true
		}
	}
	return false
}

func deepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func createManifestCollection(generatedManifests []string) (ManifestCollection, error) {
	result := make(map[ManifestPath]ManifestContent, len(generatedManifests))

	for _, m := range generatedManifests {
		metadata, err := unmarshalMetadata([]byte(m))
		if err != nil {
			return nil, errors.Wrap(err, "while getting metadata for manifest")
		}
		manifestPath := ManifestPath(fmt.Sprintf("%s.%s", *metadata.Metadata.Prefix, metadata.Metadata.Name))
		result[manifestPath] = []byte(m)
	}

	return result, nil
}

func createFilePathFromManifestPath(path ManifestPath) string {
	return strings.ReplaceAll(strings.TrimPrefix(string(path), "cap."), ".", string(os.PathSeparator)) + ".yaml"
}

func getDefaultInputTypeName(name string) string {
	return name + "-input"
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
