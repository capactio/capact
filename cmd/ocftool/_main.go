package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
)

var rootDir = flag.String("root-dir", "", "Root dir for yaml files")

func main() {
	flag.Parse()
	files, err := WalkMatch(*rootDir, ".yaml")
	exitOnError(err)
	for _, name := range files {
		fmt.Println(name)
		dat, err := ioutil.ReadFile(name)
		exitOnError(err)

		out, err := yaml.YAMLToJSON(dat)
		exitOnError(err)

		outname := "./pkg/apis/0.0.1/" + strings.TrimSuffix(filepath.Base(name), filepath.Ext(name)) + ".json"
		err = ioutil.WriteFile(outname, out, 0644)
		exitOnError(err)
	}
}

func exitOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}

}

func WalkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext == pattern {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}
