package helmstoragebackend_test

import (
	"context"
	"testing"

	helmstoragebackend "capact.io/capact/internal/helm-storage-backend"
	hublocalgraphql "capact.io/capact/pkg/hub/api/graphql/local"
	"capact.io/capact/pkg/hub/client/local"
	helmrunner "capact.io/capact/pkg/runner/helm"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKubeconfigFetcher_FetchByTypeInstanceID(t *testing.T) {
	// given
	id := "foo"
	fetchErrMsg := "while fetching TypeInstance \"foo\": sample err"
	unmarshalErrMsg := "while unmarshalling TypeInstance \"foo\" into kubeconfig: json: cannot unmarshal string into Go struct field KubeconfigInput.Value of type helm.KubeconfigContent"
	tiNotFoundErrMsg := "while getting TypeInstance: TypeInstance with IO \"foo\" not found"
	ti := map[string]interface{}{
		"config": map[string]interface{}{
			"testkey": "testvalue",
		},
	}

	kubeconfig := helmrunner.KubeconfigInput{
		Value: helmrunner.KubeconfigContent{
			Config: map[string]interface{}{
				"testkey": "testvalue",
			},
		},
	}

	testCases := []struct {
		Name               string
		TIGetter           helmstoragebackend.TypeInstanceGetter
		Expected           helmrunner.KubeconfigInput
		ExpectedErrMessage *string
	}{
		{
			Name:     "Success",
			TIGetter: &fakeTIGetter{expectedID: id, valueToReturn: ti},
			Expected: kubeconfig,
		},
		{
			Name:               "TI Getter Error",
			TIGetter:           &fakeTIGetter{expectedID: id, errToReturn: errors.New("sample err")},
			ExpectedErrMessage: &fetchErrMsg,
		},
		{
			Name:               "Unmarshalling Error",
			TIGetter:           &fakeTIGetter{expectedID: id, valueToReturn: "invalid"},
			ExpectedErrMessage: &unmarshalErrMsg,
		},
		{
			Name:               "TypeInstance not existing",
			TIGetter:           &fakeTIGetter{expectedID: id, returnNil: true},
			ExpectedErrMessage: &tiNotFoundErrMsg,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			fetcher := helmstoragebackend.NewKubeconfigFetcher(testCase.TIGetter)

			// when
			res, err := fetcher.FetchByTypeInstanceID(context.Background(), id)

			// then
			if testCase.ExpectedErrMessage != nil {
				require.Error(t, err)
				assert.EqualError(t, err, *testCase.ExpectedErrMessage)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.Expected, res)
			}
		})
	}
}

type fakeTIGetter struct {
	expectedID    string
	valueToReturn interface{}
	errToReturn   error
	returnNil     bool
}

func (f fakeTIGetter) FindTypeInstance(ctx context.Context, id string, opts ...local.TypeInstancesOption) (*hublocalgraphql.TypeInstance, error) {
	if id != f.expectedID {
		return nil, errors.New("invalid id")
	}

	if f.errToReturn != nil {
		return nil, f.errToReturn
	}

	if f.returnNil {
		return nil, nil
	}

	return &hublocalgraphql.TypeInstance{
		ID:       f.expectedID,
		LockedBy: nil,
		LatestResourceVersion: &hublocalgraphql.TypeInstanceResourceVersion{
			Spec: &hublocalgraphql.TypeInstanceResourceVersionSpec{
				Value: f.valueToReturn,
			},
		},
	}, nil
}
