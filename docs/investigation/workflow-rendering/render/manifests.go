package render

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
	"sigs.k8s.io/yaml"
)

type ManifestStore struct {
	ManifestDir     string
	Implementations map[v1alpha1.ManifestReference]*types.Implementation
	Interfaces      map[v1alpha1.ManifestReference]*types.Interface
	TypeInstance    map[string]*types.TypeInstance
}

func NewManifestStore(manifestDir string) (*ManifestStore, error) {
	store := &ManifestStore{
		ManifestDir:     manifestDir,
		Implementations: map[v1alpha1.ManifestReference]*types.Implementation{},
		Interfaces:      map[v1alpha1.ManifestReference]*types.Interface{},
		TypeInstance:    map[string]*types.TypeInstance{},
	}

	err := store.loadManifests()
	if err != nil {
		return nil, errors.Wrap(err, "while loading manifests")
	}

	return store, err
}

func (s *ManifestStore) GetImplementation(ref v1alpha1.ManifestReference) *types.Implementation {
	return s.Implementations[ref]
}

func (s *ManifestStore) GetImplementationForInterface(actionPath string) *types.Implementation {
	for key := range s.Implementations {
		impl := s.Implementations[key]

		for _, implements := range impl.Spec.Implements {
			if implements.Path == actionPath {
				return impl
			}
		}
	}

	return nil
}

func (s *ManifestStore) GetInterface(ref v1alpha1.ManifestReference) *types.Interface {
	return s.Interfaces[ref]
}

func (s *ManifestStore) GetTypeInstance(ID string) *types.TypeInstance {
	return s.TypeInstance[ID]
}

type manifestMetadata struct {
	Kind string `json:"kind"`
}

func (s *ManifestStore) loadManifests() error {
	err := filepath.Walk(s.ManifestDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
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

func (s *ManifestStore) loadManifest(filepath string) error {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return errors.Wrap(err, "while reading file")
	}

	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return errors.Wrap(err, "while converting YAML to JSON")
	}

	manifest := &manifestMetadata{}
	if err := json.Unmarshal(jsonData, manifest); err != nil {
		return errors.Wrap(err, "while unmarshaling manifestMetadata")
	}

	switch manifest.Kind {
	case "Implementation":
		impl, err := types.UnmarshalImplementation(jsonData)
		if err != nil {
			return errors.Wrap(err, "while unmarshaling Implementation")
		}

		key := fmt.Sprintf("%s.%s", *impl.Metadata.Prefix, impl.Metadata.Name)
		s.Implementations[v1alpha1.ManifestReference{
			Path: v1alpha1.NodePath(key),
		}] = &impl

	case "Interface":
		iface, err := types.UnmarshalInterface(jsonData)
		if err != nil {
			return errors.Wrap(err, "while unmarshaling Interface")
		}

		key := fmt.Sprintf("%s.%s", *iface.Metadata.Prefix, iface.Metadata.Name)
		s.Interfaces[v1alpha1.ManifestReference{
			Path: v1alpha1.NodePath(key),
		}] = &iface

	case "TypeInstance":
		typeInstance, err := types.UnmarshalTypeInstance(jsonData)
		if err != nil {
			return errors.Wrap(err, "while unmarshaling TypeInstance")
		}

		s.TypeInstance[typeInstance.Metadata.ID] = &typeInstance
	}

	return nil
}
