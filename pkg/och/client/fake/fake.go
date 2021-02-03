package fake

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	ochlocalgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	ochpublicgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/public"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

const manifestsExtension = ".yaml"

type FileSystemClient struct {
	OCHTypeInstances   map[string]ochlocalgraphql.TypeInstance
	OCHImplementations []ochpublicgraphql.ImplementationRevision
}

func NewFromLocal(manifestDir string) (*FileSystemClient, error) {
	store := &FileSystemClient{
		OCHImplementations: []ochpublicgraphql.ImplementationRevision{},
		OCHTypeInstances:   map[string]ochlocalgraphql.TypeInstance{},
	}

	if err := store.loadManifests(manifestDir); err != nil {
		return nil, errors.Wrap(err, "while loading OCH manifests")
	}

	return store, nil
}

func (s *FileSystemClient) GetImplementationForInterface(_ context.Context, ref ochpublicgraphql.TypeReference) (*ochpublicgraphql.ImplementationRevision, error) {
	for _, impl := range s.OCHImplementations {
		for _, implements := range impl.Spec.Implements {
			if implements.Path == ref.Path {
				return &impl, nil
			}
		}
	}

	return nil, fmt.Errorf("implementation for %v not found", ref)
}

func (s *FileSystemClient) GetTypeInstance(_ context.Context, id string) (*ochlocalgraphql.TypeInstance, error) {
	ti, found := s.OCHTypeInstances[id]
	if !found {
		return nil, fmt.Errorf("type instance with id %v not found", id)
	}

	return &ti, nil
}

func (s *FileSystemClient) loadManifests(dir string) error {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if ext := filepath.Ext(path); ext != manifestsExtension {
			return nil
		}

		if err := s.loadManifest(path); err != nil {
			return errors.Wrapf(err, "while loading manifest %s", path)
		}

		return nil
	})

	if err != nil {
		return errors.Wrap(err, "while walking through manifest dir")
	}

	return nil
}

func (s *FileSystemClient) loadManifest(filepath string) error {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return errors.Wrap(err, "while reading file")
	}

	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return errors.Wrap(err, "while converting YAML to JSON")
	}

	if strings.Contains(filepath, "implementation") {
		impl := ochpublicgraphql.ImplementationRevision{}
		if err := json.Unmarshal(jsonData, &impl); err != nil {
			return err
		}
		s.OCHImplementations = append(s.OCHImplementations, impl)
	}

	if strings.Contains(filepath, "typeinstance") {
		ti := ochlocalgraphql.TypeInstance{}
		if err := json.Unmarshal(jsonData, &ti); err != nil {
			return err
		}
		s.OCHTypeInstances[ti.Metadata.ID] = ti
	}

	return nil
}
