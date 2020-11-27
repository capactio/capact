package cloudsql

import (
	"fmt"
	"html/template"
	"os"

	"github.com/pkg/errors"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"sigs.k8s.io/yaml"
)

type Output struct {
	DBInstance    *sqladmin.DatabaseInstance
	Port          int    `json:"port"`
	DefaultDBName string `json:"defaultDBName"`
	Username      string `json:"username"`
	Password      string `json:"password"`
}

const cloudSQLInstanceOutputTemplate string = `
name: "{{ .DBInstance.Name }}"
project: "{{ .DBInstance.Project }}"
region: "{{ .DBInstance.Region }}"
databaseVersion: "{{ .DBInstance.DatabaseVersion }}"
`

func writeOutput(args *OutputArgs, output *Output) error {
	yamlBytes, err := yaml.JSONToYAML(args.Additional.Value)
	if err != nil {
		return errors.Wrap(err, "cannot convert output json to yaml")
	}

	if err := writeOutputFile(args.Directory, args.Default.Filename, string(cloudSQLInstanceOutputTemplate), output); err != nil {
		return errors.Wrap(err, "failed to write default output file")
	}

	if err := writeOutputFile(args.Directory, args.Additional.Path, string(yamlBytes), output); err != nil {
		return errors.Wrap(err, "failed to write additional output file")
	}
	return nil
}

func writeOutputFile(outDir, outFilename, templateStr string, values interface{}) error {
	tmpl, err := template.New("output").Parse(templateStr)
	if err != nil {
		return errors.Wrap(err, "failed to load template")
	}

	if err := os.MkdirAll(outDir, 0775); err != nil {
		return err
	}

	filepath := fmt.Sprintf("%s/%s", outDir, outFilename)

	fd, err := os.Create(filepath)
	if err != nil {
		return errors.Wrap(err, "cannot open output file to write")
	}
	defer fd.Close()

	err = tmpl.Execute(fd, values)
	if err != nil {
		return errors.Wrap(err, "failed to render output file")
	}

	return nil
}
