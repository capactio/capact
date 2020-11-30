package cloudsql

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"sigs.k8s.io/yaml"
)

type outputValues struct {
	DBInstance    *sqladmin.DatabaseInstance
	Port          int    `json:"port"`
	DefaultDBName string `json:"defaultDBName"`
	Username      string `json:"username"`
	Password      string `json:"password"`
}

type instanceArtifact struct {
	Name            string `json:"name"`
	Project         string `json:"project"`
	Region          string `json:"region"`
	DatabaseVersion string `json:"databaseVersion"`
}

func createArtifacts(args *OutputArgs, values *outputValues) error {
	if err := os.MkdirAll(args.Directory, 0775); err != nil {
		return err
	}

	if err := createCloudSQLInstanceArtifact(args, values); err != nil {
		return errors.Wrap(err, "while creating default artifact")
	}

	if args.Additional != nil {
		if err := createAdditionalArtifact(args, values); err != nil {
			return errors.Wrap(err, "while creating additional artifact")
		}
	}

	return nil
}

func createCloudSQLInstanceArtifact(args *OutputArgs, output *outputValues) error {
	artifact := &instanceArtifact{
		Name:            output.DBInstance.Name,
		Project:         output.DBInstance.Project,
		Region:          output.DBInstance.Region,
		DatabaseVersion: output.DBInstance.DatabaseVersion,
	}

	data, err := yaml.Marshal(artifact)
	if err != nil {
		return errors.Wrap(err, "while marshaling artifact to YAML")
	}

	artifactFilepath := fmt.Sprintf("%s/%s", args.Directory, args.CloudSQLInstance.Filename)

	// #nosec G306: Poor file permissions used when writing to a new file
	if err := ioutil.WriteFile(artifactFilepath, data, 0644); err != nil {
		return errors.Wrapf(err, "while writing artifact file %s", artifactFilepath)
	}

	return nil
}

func createAdditionalArtifact(args *OutputArgs, values *outputValues) error {
	artifactTemplate, err := yaml.JSONToYAML(args.Additional.Value)
	if err != nil {
		return errors.Wrap(err, "while converting JSON to YAML")
	}

	tmpl, err := template.New("output").Parse(string(artifactTemplate))
	if err != nil {
		return errors.Wrap(err, "failed to load template")
	}

	filepath := fmt.Sprintf("%s/%s", args.Directory, args.Additional.Path)

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
