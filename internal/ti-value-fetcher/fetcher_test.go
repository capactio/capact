package tivaluefetcher_test

import (
	"context"
	"encoding/json"
	"net"
	"testing"

	"capact.io/capact/internal/logger"
	"capact.io/capact/internal/ptr"
	tivaluefetcher "capact.io/capact/internal/ti-value-fetcher"
	"capact.io/capact/pkg/hub/api/grpc/storage_backend"
	storagebackend "capact.io/capact/pkg/hub/storage-backend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestTIValueFetcher_LoadFromFile_HappyPath(t *testing.T) {
	// given
	inputTIPath := "testdata/input-ti.yaml"
	storageBackendPath := "testdata/storage-backend.yaml"
	expectedTIData := tivaluefetcher.TypeInstanceData{
		Value: nil,
		Backend: &tivaluefetcher.Backend{
			Context: []byte(`{"chartLocation":"https://charts.bitnami.com/bitnami","driver":"secrets","name":"example-release","namespace":"default"}`),
		},
	}
	expectedStorageBackendData := storagebackend.TypeValue{
		URL:         "localhost:50051",
		AcceptValue: false,
		ContextSchema: map[string]interface{}{
			"$schema":              "http://json-schema.org/draft-07/schema",
			"additionalProperties": false,
			"properties": map[string]interface{}{
				"chartLocation": map[string]interface{}{
					"$id":  "#/properties/context/properties/chartLocation",
					"type": "string",
				},
				"driver": map[string]interface{}{
					"$id":     "#/properties/context/properties/driver",
					"default": "secrets",
					"enum": []interface{}{
						"secrets",
						"configmaps",
						"sql",
					},
					"type": "string",
				},
				"name": map[string]interface{}{
					"$id":  "#/properties/context/properties/name",
					"type": "string",
				},
				"namespace": map[string]interface{}{
					"$id":  "#/properties/context/properties/namespace",
					"type": "string"},
			},
			"required": []interface{}{
				"name",
				"namespace",
				"chartLocation",
			},
			"type": "object",
		},
	}

	log := logger.Noop()
	tiValueFetcher := tivaluefetcher.New(log)

	// when
	tiData, storageBackend, err := tiValueFetcher.LoadFromFile(inputTIPath, storageBackendPath)

	// then
	require.NoError(t, err)
	assert.Equal(t, expectedTIData, tiData)
	assert.Equal(t, expectedStorageBackendData, storageBackend)
}

func TestTIValueFetcher_Do(t *testing.T) {
	// given
	log := logger.Noop()
	sampleValue := json.RawMessage(`foo: bar`)

	for _, testCase := range []struct {
		Name                     string
		InputTypeInstanceData    tivaluefetcher.TypeInstanceData
		InputStorageBackendValue storagebackend.TypeValue
		BackendServerHandlerFn   func(t *testing.T) storage_backend.ContextStorageBackendServer
		ExpectedResult           tivaluefetcher.TypeInstanceData
		ExpectedErrorMessage     *string
	}{
		{
			Name: "Skip getting value if it is already there",
			InputTypeInstanceData: tivaluefetcher.TypeInstanceData{
				Value: &sampleValue,
				Backend: &tivaluefetcher.Backend{
					Context: []byte(`sample: context`),
				},
			},
			ExpectedResult: tivaluefetcher.TypeInstanceData{
				Value: &sampleValue,
				Backend: &tivaluefetcher.Backend{
					Context: []byte(`sample: context`),
				},
			},
		},
		{
			Name: "Skip getting value if backend accepts value",
			InputTypeInstanceData: tivaluefetcher.TypeInstanceData{
				Value: nil,
				Backend: &tivaluefetcher.Backend{
					Context: []byte(`sample: context`),
				},
			},
			InputStorageBackendValue: storagebackend.TypeValue{
				URL:           "foo.bar",
				AcceptValue:   true,
				ContextSchema: map[string]interface{}{},
			},
			ExpectedResult: tivaluefetcher.TypeInstanceData{
				Value: nil,
				Backend: &tivaluefetcher.Backend{
					Context: []byte(`sample: context`),
				},
			},
		},
		{
			Name: "Get value from backend",
			InputTypeInstanceData: tivaluefetcher.TypeInstanceData{
				Value: nil,
				Backend: &tivaluefetcher.Backend{
					Context: []byte(`sample: context`),
				},
			},
			InputStorageBackendValue: storagebackend.TypeValue{
				URL:           "foo.bar",
				AcceptValue:   false,
				ContextSchema: map[string]interface{}{},
			},
			BackendServerHandlerFn: func(t *testing.T) storage_backend.ContextStorageBackendServer {
				return &fakeHandler{
					t:           t,
					expectedCtx: []byte(`sample: context`),
					value:       sampleValue,
				}
			},
			ExpectedResult: tivaluefetcher.TypeInstanceData{
				Value: &sampleValue,
				Backend: &tivaluefetcher.Backend{
					Context: []byte(`sample: context`),
				},
			},
		},
		{
			Name: "Error from backend",
			InputTypeInstanceData: tivaluefetcher.TypeInstanceData{
				Value: nil,
				Backend: &tivaluefetcher.Backend{
					Context: []byte(`sample: context`),
				},
			},
			InputStorageBackendValue: storagebackend.TypeValue{
				URL:           "foo.bar",
				AcceptValue:   false,
				ContextSchema: map[string]interface{}{},
			},
			BackendServerHandlerFn: func(t *testing.T) storage_backend.ContextStorageBackendServer {
				return &fakeHandler{
					t:           t,
					expectedCtx: []byte(`sample: context`),
					err:         status.Error(codes.Internal, "sample error"),
				}
			},
			ExpectedErrorMessage: ptr.String("while getting precreate value from storage backend \"foo.bar\": rpc error: code = Internal desc = sample error"),
		},
	} {
		t.Run(testCase.Name, func(t *testing.T) {
			var dialOpts []grpc.DialOption

			if testCase.BackendServerHandlerFn != nil {
				handler := testCase.BackendServerHandlerFn(t)
				srv, listener := setupFakeServerAndListener(t, handler)
				defer srv.Stop()
				dialOpts = dialOptsForListener(listener)
			}

			tiValueFetcher := tivaluefetcher.New(log)

			// when
			result, err := tiValueFetcher.Do(
				context.Background(),
				testCase.InputTypeInstanceData,
				testCase.InputStorageBackendValue,
				dialOpts...,
			)

			// then
			if testCase.ExpectedErrorMessage != nil {
				require.Error(t, err)
				assert.EqualError(t, err, *testCase.ExpectedErrorMessage)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, testCase.ExpectedResult, result)
		})
	}
}

const bufSize = 1024 * 1024

func setupFakeServerAndListener(t *testing.T, handler storage_backend.ContextStorageBackendServer) (*grpc.Server, *bufconn.Listener) {
	t.Helper()

	listener := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	storage_backend.RegisterContextStorageBackendServer(srv, handler)

	go func() {
		err := srv.Serve(listener)
		require.NoError(t, err)
	}()

	return srv, listener
}

func dialOptsForListener(listener *bufconn.Listener) []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithInsecure(),
	}
}

type fakeHandler struct {
	storage_backend.UnimplementedContextStorageBackendServer
	t           *testing.T
	expectedCtx []byte
	err         error
	value       []byte
}

func (h *fakeHandler) GetPreCreateValue(_ context.Context, req *storage_backend.GetPreCreateValueRequest) (*storage_backend.GetPreCreateValueResponse, error) {
	if h.err != nil {
		return nil, h.err
	}

	require.NotNil(h.t, req)
	assert.Equal(h.t, h.expectedCtx, req.Context)

	return &storage_backend.GetPreCreateValueResponse{
		Value: h.value,
	}, nil
}
