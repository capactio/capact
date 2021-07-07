package io

import (
	"os"
	"path/filepath"
	"strings"
)

// List returns all YAML files in the provided path by using filepath.Walk.
func ListYamls(path string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(path, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && isYaml(info.Name()) {
			files = append(files, currentPath)
		}
		return nil
	})
	return files, err
}

func isYaml(path string) bool {
	return strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")
}
