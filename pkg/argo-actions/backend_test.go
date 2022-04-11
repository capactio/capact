package argoactions

import (
	"context"
	"testing"

	"capact.io/capact/internal/ptr"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	"capact.io/capact/pkg/hub/client/local"
	storagebackend "capact.io/capact/pkg/hub/storage-backend"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveBackendsValues(t *testing.T) {
	// given
	ids := []string{"foo", "bar"}
	fooVal := map[string]interface{}{
		"acceptValue":   true,
		"url":           "foo.baz",
		"contextSchema": nil,
	}
	barVal := map[string]interface{}{
		"acceptValue": false,
		"url":         "bar.baz",
		"contextSchema": map[string]interface{}{
			"$schema": "http://json-schema.org/draft-07/schema#",
		},
	}
	expectedFooBackendValue := storagebackend.TypeInstanceValue{
		URL:           "foo.baz",
		AcceptValue:   true,
		ContextSchema: nil,
	}
	expectedBarBackendValue := storagebackend.TypeInstanceValue{
		URL:         "bar.baz",
		AcceptValue: false,
		ContextSchema: map[string]interface{}{
			"$schema": "http://json-schema.org/draft-07/schema#",
		},
	}

	testCases := []struct {
		Name               string
		Client             findTIClient
		ExpectedErrMessage *string
		ExpectedResult     map[string]storagebackend.TypeInstanceValue
	}{
		{
			Name: "Success",
			Client: &fakeFindTIClient{
				result: map[string]gqllocalapi.TypeInstance{
					"foo": fixTI(fooVal),
					"bar": fixTI(barVal),
				},
			},
			ExpectedResult: map[string]storagebackend.TypeInstanceValue{
				"foo": expectedFooBackendValue,
				"bar": expectedBarBackendValue,
			},
		},
		{
			Name: "Error",
			Client: &fakeFindTIClient{
				err: errors.New("test error"),
			},
			ExpectedErrMessage: ptr.String("while fetching TypeInstance values: test error"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// when
			res, err := resolveBackendsValues(context.Background(), testCase.Client, ids)

			// then
			if testCase.ExpectedErrMessage != nil {
				require.Error(t, err)
				assert.EqualError(t, err, *testCase.ExpectedErrMessage)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, res)
			}
		})
	}
}

func fixTI(value interface{}) gqllocalapi.TypeInstance {
	return gqllocalapi.TypeInstance{
		LatestResourceVersion: &gqllocalapi.TypeInstanceResourceVersion{
			Spec: &gqllocalapi.TypeInstanceResourceVersionSpec{
				Value: value,
			},
		}}
}

type fakeFindTIClient struct {
	result map[string]gqllocalapi.TypeInstance
	err    error
}

func (f *fakeFindTIClient) FindTypeInstances(ctx context.Context, _ []string, _ ...local.TypeInstancesOption) (map[string]gqllocalapi.TypeInstance, error) {
	return f.result, f.err
}
