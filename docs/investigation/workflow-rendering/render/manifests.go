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

const manifestsExtension = ".yaml"

type ManifestStore struct {
	Implementations map[v1alpha1.ManifestReference]*types.Implementation
	Interfaces      map[v1alpha1.ManifestReference]*types.Interface
	TypeInstance    map[string]*types.TypeInstance
}

func NewManifestStore(manifestDir string, typeInstanceDir string) (*ManifestStore, error) {
	store := &ManifestStore{
		Implementations: map[v1alpha1.ManifestReference]*types.Implementation{},
		Interfaces:      map[v1alpha1.ManifestReference]*types.Interface{},
		TypeInstance:    map[string]*types.TypeInstance{},
	}

	if err := store.loadManifests(manifestDir); err != nil {
		return nil, errors.Wrap(err, "while loading OCH manifests")
	}

	if err := store.loadManifests(typeInstanceDir); err != nil {
		return nil, errors.Wrap(err, "while loading TypeInstances manifests")
	}

	return store, nil
}

func (s *ManifestStore) GetImplementation(ref v1alpha1.ManifestReference) *types.Implementation {
	return s.Implementations[ref]
}

type RequiresFilter struct {
	Excluded []string `json:"excluded"`
}

type GetImplementationForInterfaceInput struct {
	RequireFilter RequiresFilter
}

func (s *ManifestStore) GetImplementationForInterface(actionPath string, in GetImplementationForInterfaceInput) *types.Implementation {
	for key := range s.Implementations {
		impl := s.Implementations[key]

		if !requirementsAreSatisfied(impl.Spec.Requires, in) {
			continue
		}
		for _, implements := range impl.Spec.Implements {

			if implements.Path == actionPath {
				return impl
			}
		}
	}

	return nil
}

// TODO: dummy impl, needs to be fixed.
func requirementsAreSatisfied(requires map[string]types.Require, in GetImplementationForInterfaceInput) bool {
	for _, req := range requires {

		if len(req.AllOf) > 0 {
			aReq := filterOut(req.AllOf, in.RequireFilter.Excluded)
			if len(aReq) != len(req.AllOf) { // sth was excluded by filter but need by impl
				return false
			}
		}
		if len(req.OneOf) > 0 {
			oReq := filterOut(req.OneOf, in.RequireFilter.Excluded)
			if len(oReq) < 1 { // everything was filter out but at least one needs to stay
				return false
			}
		}

		if len(req.AnyOf) > 0 {
			oReq := filterOut(req.AnyOf, in.RequireFilter.Excluded)
			if len(oReq) < 1 { // everything was filter out but at least one needs to stay
				return false
			}
		}
	}

	return true
}

func filterOut(req []types.RequireEntity, excluded []string) []types.RequireEntity {
	if len(excluded) == 0 {
		return req
	}

	toExclude := map[string]struct{}{}
	for _, v := range excluded {
		toExclude[v] = struct{}{}
	}

	var out []types.RequireEntity
	for _, aReq := range req {
		if _, exclude := toExclude[aReq.Name]; exclude {
			continue
		}
		out = append(out, aReq)
	}
	return out
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

func (s *ManifestStore) loadManifests(dir string) error {
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
