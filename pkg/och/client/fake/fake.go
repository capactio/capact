package fake

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"

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
	getOpts := &public.GetImplementationOptions{}
	getOpts.Apply(opts...)

	var out []ochpublicgraphql.ImplementationRevision
	for i := range s.OCHImplementations {
		impl := s.OCHImplementations[i]
		for _, implements := range impl.Spec.Implements {
			if implements.Path == ref.Path {
				item := ochpublicgraphql.ImplementationRevision{}

				if err := deepCopy(&impl, &item); err != nil {
					return nil, err
				}

				out = append(out, item)
			}
		}
	}

	result := public.FilterImplementationRevisions(out, getOpts)
	if len(result) == 0 {
		return nil, public.NewImplementationRevisionNotFoundError(ref)
	}

	return result, nil
}

func (s *FileSystemClient) ListTypeInstancesTypeRef(ctx context.Context) ([]ochlocalgraphql.TypeReference, error) {
	var typeInstanceTypeRefs []ochlocalgraphql.TypeReference
	for _, ti := range s.OCHTypeInstances {
		if ti.Spec == nil || ti.Spec.TypeRef == nil {
			continue
		}

		typeInstanceTypeRefs = append(typeInstanceTypeRefs, *ti.Spec.TypeRef)
	}

	return typeInstanceTypeRefs, nil
}

func (s *FileSystemClient) GetInterfaceLatestRevisionString(ctx context.Context, ref ochpublicgraphql.InterfaceReference) (string, error) {
	var versions semver.Collection
	for _, impl := range s.OCHImplementations {
		for _, implements := range impl.Spec.Implements {
			if implements.Path == ref.Path {
				v, err := semver.NewVersion(implements.Revision)
				if err != nil {
					return "", err
				}
				versions = append(versions, v)
			}
		}
	}

	if len(versions) == 0 {
		return "", errors.New("no Interface found for a given ref")
	}

	sort.Sort(versions)
	latestVersion := versions[len(versions)-1]
	return latestVersion.String(), nil
}

func (s *FileSystemClient) GetInterfaceRevision(ctx context.Context, ref ochpublicgraphql.InterfaceReference) (*ochpublicgraphql.InterfaceRevision, error) {
	for i := range s.OCHInterfaces {
		iface := s.OCHInterfaces[i]
		if iface.Metadata.Path != ref.Path {
			continue
		}

		item := ochpublicgraphql.InterfaceRevision{}

		if err := deepCopy(&iface, &item); err != nil {
			return nil, err
		}

		return &item, nil
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

func deepCopy(src interface{}, dst interface{}) error {
	var mod bytes.Buffer
	enc := gob.NewEncoder(&mod)
	dec := gob.NewDecoder(&mod)

	err := enc.Encode(src)
	if err != nil {
		return err
	}

	return dec.Decode(dst)
}
