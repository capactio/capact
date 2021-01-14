package mockedresolver

import (
	"encoding/json"
	"io/ioutil"
	"path"

	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

const MocksPath = "./mock/public"

func MockedInterfaces() ([]*gqlpublicapi.Interface, error) {
	buff, err := ioutil.ReadFile(path.Join(MocksPath, "interfaces.json"))
	if err != nil {
		return nil, err
	}

	i := []*gqlpublicapi.Interface{}
	err = json.Unmarshal(buff, &i)
	if err != nil {
		return nil, err
	}
	for _, iface := range i {
		if len(iface.Revisions) > 0 {
			iface.LatestRevision = iface.Revisions[0]
		}
	}
	return i, nil
}

func MockedImplementations() ([]*gqlpublicapi.Implementation, error) {
	buff, err := ioutil.ReadFile(path.Join(MocksPath, "implementations.json"))
	if err != nil {
		return nil, err
	}

	i := []*gqlpublicapi.Implementation{}
	err = json.Unmarshal(buff, &i)
	if err != nil {
		return nil, err
	}
	for _, implementation := range i {
		if len(implementation.Revisions) > 0 {
			implementation.LatestRevision = implementation.Revisions[0]
		}
	}
	return i, nil
}

func MockedTypes() ([]*gqlpublicapi.Type, error) {
	buff, err := ioutil.ReadFile(path.Join(MocksPath, "types.json"))
	if err != nil {
		return nil, err
	}

	types := []*gqlpublicapi.Type{}
	err = json.Unmarshal(buff, &types)
	if err != nil {
		return nil, err
	}
	for _, t := range types {
		t.LatestRevision = t.Revisions[0]
	}
	return types, nil
}

func MockedAttributes() ([]*gqlpublicapi.Attribute, error) {
	buff, err := ioutil.ReadFile(path.Join(MocksPath, "attributes.json"))
	if err != nil {
		return nil, err
	}

	attributes := []*gqlpublicapi.Attribute{}
	err = json.Unmarshal(buff, &attributes)
	if err != nil {
		return nil, err
	}
	for _, attribute := range attributes {
		attribute.LatestRevision = attribute.Revisions[0]
	}
	return attributes, nil
}

func MockedRepoMetadata() (*gqlpublicapi.RepoMetadata, error) {
	buff, err := ioutil.ReadFile(path.Join(MocksPath, "repoMetadata.json"))
	if err != nil {
		return nil, err
	}

	repo := &gqlpublicapi.RepoMetadata{}
	err = json.Unmarshal(buff, repo)
	if err != nil {
		return nil, err
	}
	repo.LatestRevision = repo.Revisions[0]
	return repo, nil
}

func MockedInterfaceGroups() ([]*gqlpublicapi.InterfaceGroup, error) {
	buff, err := ioutil.ReadFile(path.Join(MocksPath, "interfaceGroups.json"))
	if err != nil {
		return nil, err
	}

	groups := []*gqlpublicapi.InterfaceGroup{}
	err = json.Unmarshal(buff, &groups)
	if err != nil {
		return nil, err
	}
	return groups, nil
}
