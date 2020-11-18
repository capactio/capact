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
