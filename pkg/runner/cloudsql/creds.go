package cloudsql

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
)

type Config struct {
	ServiceAccount struct {
		Filepath   string            `envconfig:"default=/etc/gcp/sa.json"`
		FileFormat CredentialsFormat `envconfig:"default=json"`
	}
}

// CredentialsFormat represents the possible credentials format. It works with the envconfig library.
type CredentialsFormat string

const (
	JSON CredentialsFormat = "JSON"
	YAML CredentialsFormat = "YAML"
)

// Validate returns errors if Mode is unknown.
func (h CredentialsFormat) Validate() error {
	switch h {
	case JSON, YAML:
		return nil
	}
	return fmt.Errorf("Wrong credentials format. Possible options: %s and %s", JSON, YAML)
}

// Unmarshal fulfils the envconfig interface for unmarshaling.
func (h *CredentialsFormat) Unmarshal(s string) error {
	hub := CredentialsFormat(s)
	if err := hub.Validate(); err != nil {
		return err
	}
	*h = hub
	return nil
}

var scopes = []string{
	"https://www.googleapis.com/auth/cloud-platform",
	"https://www.googleapis.com/auth/sqlservice.admin",
}

func LoadGCPCredentials(cfg Config) (*google.Credentials, error) {
	rawInput, err := ioutil.ReadFile(cfg.ServiceAccount.Filepath)
	if err != nil {
		return nil, errors.Wrap(err, "while reading GCP credentials file")
	}

	switch cfg.ServiceAccount.FileFormat {
	case YAML:
		rawInput, err = yaml.YAMLToJSON(rawInput)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert credentials from YAML to JSON")
		}
	}

	gcpCreds, err := google.CredentialsFromJSON(context.Background(), rawInput, scopes...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read GCP credentials")
	}

	return gcpCreds, nil
}
