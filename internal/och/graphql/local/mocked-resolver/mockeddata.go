package mockedresolver

import (
	"encoding/json"
	"io/ioutil"
	"path"

	gqllocalapi "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
)

const MocksPath = "./mock/local"

func MockedTypeInstances() ([]*gqllocalapi.TypeInstance, error) {
	buff, err := ioutil.ReadFile(path.Join(MocksPath, "typeInstances.json"))
	if err != nil {
		return nil, err
	}

	i := []*gqllocalapi.TypeInstance{}
	err = json.Unmarshal(buff, &i)
	if err != nil {
		return nil, err
	}
	return i, nil
}
