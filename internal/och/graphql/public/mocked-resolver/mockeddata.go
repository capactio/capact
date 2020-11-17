package mockedresolver

import (
	"encoding/json"
	"io/ioutil"

	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

func MockedInterface() (*gqlpublicapi.Interface, error) {
	i := &gqlpublicapi.Interface{}

	buff, err := ioutil.ReadFile("interface.json")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(buff, i)
	if err != nil {
		return nil, err
	}
	i.LatestRevision = i.Revisions[0]
	return i, nil
}
