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
	"projectvoltron.dev/voltron/pkg/och/client/public"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

const manifestsExtension = ".yaml"

type FileSystemClient struct {
	OCHTypeInstances   map[string]ochlocalgraphql.TypeInstance
	OCHImplementations []ochpublicgraphql.ImplementationRevision
	OCHInterfaces      []ochpublicgraphql.InterfaceRevision
}

func NewFromLocal(manifestDir string) (*FileSystemClient, error) {
	store := &FileSystemClient{
		OCHImplementations: []ochpublicgraphql.ImplementationRevision{},
		OCHInterfaces:      []ochpublicgraphql.InterfaceRevision{},
		OCHTypeInstances:   map[string]ochlocalgraphql.TypeInstance{},
	}

	if err := store.loadManifests(manifestDir); err != nil {
		return nil, errors.Wrap(err, "while loading OCH manifests")
	}

	return store, nil
}

func (s *FileSystemClient) GetImplementationRevisionsForInterface(ctx context.Context, ref ochpublicgraphql.InterfaceReference, opts ...public.GetImplementationOption) ([]ochpublicgraphql.ImplementationRevision, error) {
	var out []ochpublicgraphql.ImplementationRevision
	for _, impl := range s.OCHImplementations {
		for _, implements := range impl.Spec.Implements {
			if implements.Path == ref.Path {
				out = append(out, impl)
			}
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no ImplementationRevisions for Interface %v", ref)
	}
	return out, nil
}

func (s *FileSystemClient) GetInterfaceRevision(ctx context.Context, ref ochpublicgraphql.InterfaceReference) (*ochpublicgraphql.InterfaceRevision, error) {
	for i := range s.OCHInterfaces {
		iface := s.OCHInterfaces[i]
		if *iface.Metadata.Path == ref.Path {
			return &iface, nil
		}
	}

	return nil, fmt.Errorf("cannot find Interface for %v", ref)
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

	if strings.Contains(filepath, "interface") {
		iface := ochpublicgraphql.InterfaceRevision{}
		if err := json.Unmarshal(jsonData, &iface); err != nil {
			return err
		}
		s.OCHInterfaces = append(s.OCHInterfaces, iface)
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
