// @generated - This was created as a part of investigation. We mark it as generate to exlude it from goreportcard to do not have missleading issues.:golint
package render

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"capact.io/capact/pkg/engine/k8s/api/v1alpha1"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
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

type PolicyItem struct {
	Attribute string `json:"attribute"`
}

type FilterPolicies struct {
	Included []PolicyItem `json:"included"`
	Excluded []PolicyItem `json:"excluded"`
}

type GetImplementationForInterfaceInput struct {
	Policies map[string]FilterPolicies
}

func (s *ManifestStore) GetImplementationForInterface(actionPath string, in GetImplementationForInterfaceInput) *types.Implementation {
	for key := range s.Implementations {
		impl := s.Implementations[key]

		policies, found := in.Policies[actionPath]
		if found && !policiesAreSatisfied(impl, policies) {
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

// policiesAreSatisfied verifies only Attributes
func policiesAreSatisfied(impl *types.Implementation, in FilterPolicies) bool {
	if len(in.Excluded) > 0 && containsAttributes(impl.Metadata.Attributes, in.Excluded) {
		return false
	}

	if len(in.Included) > 0 && !containsAttributes(impl.Metadata.Attributes, in.Included) {
		return false
	}

	return true
}

//  contains returns true if all items from expAtr are defined in implAtr. Duplicates are skipped.
func containsAttributes(implAtr map[string]types.MetadataAttribute, expAtr []PolicyItem) bool {
	expected := map[string]struct{}{}
	for _, v := range expAtr {
		expected[v.Attribute] = struct{}{}
	}

	matchedEntries := map[string]struct{}{}
	for atrName := range implAtr {
		if _, found := expected[atrName]; found {
			matchedEntries[atrName] = struct{}{}
		}
	}

	if len(matchedEntries) != len(expected) {
		return false
	}

	return true
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
