package mockedresolver

import (
	"encoding/json"
	"io/ioutil"
	"path"

	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

const MOCKS_PATH = "./mock/public"

func MockedInterface() (*gqlpublicapi.Interface, error) {
	buff, err := ioutil.ReadFile(path.Join(MOCKS_PATH, "interface.json"))
	if err != nil {
		return nil, err
	}

	i := &gqlpublicapi.Interface{}
	err = json.Unmarshal(buff, i)
	if err != nil {
		return nil, err
	}
	i.LatestRevision = i.Revisions[0]
	return i, nil
}

func MockedImplementation() (*gqlpublicapi.Implementation, error) {
	buff, err := ioutil.ReadFile(path.Join(MOCKS_PATH, "implementation.json"))
	if err != nil {
		return nil, err
	}

	i := &gqlpublicapi.Implementation{}
	err = json.Unmarshal(buff, i)
	if err != nil {
		return nil, err
	}
	i.LatestRevision = i.Revisions[0]
	return i, nil
}

func MockedTypes() ([]*gqlpublicapi.Type, error) {
	buff, err := ioutil.ReadFile(path.Join(MOCKS_PATH, "types.json"))
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

func MockedTag() (*gqlpublicapi.Tag, error) {
	buff, err := ioutil.ReadFile(path.Join(MOCKS_PATH, "tag.json"))
	if err != nil {
		return nil, err
	}

	tag := &gqlpublicapi.Tag{}
	err = json.Unmarshal(buff, &tag)
	if err != nil {
		return nil, err
	}
	tag.LatestRevision = tag.Revisions[0]
	return tag, nil
}

func MockedInterfaceGroup() (*gqlpublicapi.InterfaceGroup, error) {
	buff, err := ioutil.ReadFile(path.Join(MOCKS_PATH, "interfaceGroup.json"))
	if err != nil {
		return nil, err
	}

	group := &gqlpublicapi.InterfaceGroup{}
	err = json.Unmarshal(buff, &group)
	if err != nil {
		return nil, err
	}
	return group, nil
}
