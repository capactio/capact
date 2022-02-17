package secret_storage_backend_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	"capact.io/capact/internal/logger"
	"capact.io/capact/internal/ptr"
	secret_storage_backend "capact.io/capact/internal/secret-storage-backend"
	"capact.io/capact/pkg/hub/api/grpc/storage_backend"
	tellercore "github.com/spectralops/teller/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestHandler_GetValue(t *testing.T) {
	// given
	providerName := "fake"
	reqAdditionalParams := []byte(fmt.Sprintf(`{"provider":"%s"}`, providerName))
	req := &storage_backend.GetValueRequest{
		TypeinstanceId:       "uuid",
		ResourceVersion:      2,
		AdditionalParameters: reqAdditionalParams,
	}
	path := fmt.Sprintf("/capact/%s", req.TypeinstanceId)

	testCases := []struct {
		Name          string
		InputProvider tellercore.Provider
		ExpectedValue []byte
	}{
		{
			Name:          "No secret",
			InputProvider: &fakeProvider{},
			ExpectedValue: nil,
		},
		{
			Name: "Success",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {
						"2": `{"key":true}`,
					},
				},
			},
			ExpectedValue: []byte(`{"key":true}`),
		},
		{
			Name: "Empty value", // empty value is also a valid one
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {
						"2": "",
					},
				},
			},
			ExpectedValue: []byte{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			srv, listener := setupServerAndListener(t, map[string]tellercore.Provider{
				providerName: testCase.InputProvider,
			})
			defer srv.Stop()

			ctx := context.Background()
			conn, err := grpc.DialContext(ctx, "", dialOpts(listener)...)
			require.NoError(t, err)
			defer conn.Close()

			client := storage_backend.NewStorageBackendClient(conn)

			// when
			res, err := client.GetValue(ctx, req)

			// then
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Equal(t, testCase.ExpectedValue, res.Value)
		})
	}
}

func TestHandler_GetLockedBy(t *testing.T) {
	// given
	providerName := "fake"
	reqAdditionalParams := []byte(fmt.Sprintf(`{"provider":"%s"}`, providerName))
	req := &storage_backend.GetLockedByRequest{
		TypeinstanceId:       "uuid",
		AdditionalParameters: reqAdditionalParams,
	}
	path := fmt.Sprintf("/capact/%s", req.TypeinstanceId)

	testCases := []struct {
		Name             string
		InputProvider    tellercore.Provider
		ExpectedLockedBy *string
	}{
		{
			Name:             "No data",
			InputProvider:    &fakeProvider{},
			ExpectedLockedBy: nil,
		},
		{
			Name: "Empty value",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {},
				},
			},
			ExpectedLockedBy: nil,
		},
		{
			Name: "Success",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {
						"locked_by": "service/foo",
					},
				},
			},
			ExpectedLockedBy: ptr.String("service/foo"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			srv, listener := setupServerAndListener(t, map[string]tellercore.Provider{
				providerName: testCase.InputProvider,
			})
			defer srv.Stop()

			ctx := context.Background()
			conn, err := grpc.DialContext(ctx, "", dialOpts(listener)...)
			require.NoError(t, err)
			defer conn.Close()

			client := storage_backend.NewStorageBackendClient(conn)

			// when
			res, err := client.GetLockedBy(ctx, req)

			// then
			require.NoError(t, err)
			require.NotNil(t, res)

			if testCase.ExpectedLockedBy == nil {
				assert.Nil(t, res.LockedBy)
				return
			}

			require.NotNil(t, res.LockedBy)
			assert.Equal(t, *testCase.ExpectedLockedBy, *res.LockedBy)
		})
	}
}

func TestHandler_OnCreate(t *testing.T) {
	// given
	providerName := "fake"
	reqAdditionalParams := []byte(fmt.Sprintf(`{"provider":"%s"}`, providerName))
	valueBytes := []byte(`{"key": true}`)
	req := &storage_backend.OnCreateRequest{
		TypeinstanceId:       "uuid",
		Value:                valueBytes,
		AdditionalParameters: reqAdditionalParams,
	}
	path := fmt.Sprintf("/capact/%s", req.TypeinstanceId)

	testCases := []struct {
		Name                  string
		InputProvider         *fakeProvider
		ExpectedProviderState map[string]map[string]string
		ExpectedErrorMessage  *string
	}{
		{
			Name:          "No data",
			InputProvider: &fakeProvider{secrets: map[string]map[string]string{}},
			ExpectedProviderState: map[string]map[string]string{
				path: {
					"1": string(valueBytes),
				},
			},
		},
		{
			Name: "Empty value",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {},
				},
			},
			ExpectedProviderState: map[string]map[string]string{
				path: {
					"1": string(valueBytes),
				},
			},
		},
		{
			Name: "Already existing without conflict",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {
						"locked_by": "service/foo",
					},
				},
			},
			ExpectedProviderState: map[string]map[string]string{
				path: {
					"1":         string(valueBytes),
					"locked_by": "service/foo",
				},
			},
		},
		{
			Name: "Already existing with conflict",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {
						"1": "original",
					},
				},
			},
			ExpectedProviderState: map[string]map[string]string{
				path: {
					"1": "original",
				},
			},
			ExpectedErrorMessage: ptr.String("rpc error: code = AlreadyExists desc = entry \"/capact/uuid\" in provider \"fake\" already exist"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			srv, listener := setupServerAndListener(t, map[string]tellercore.Provider{
				providerName: testCase.InputProvider,
			})
			defer srv.Stop()

			ctx := context.Background()
			conn, err := grpc.DialContext(ctx, "", dialOpts(listener)...)
			require.NoError(t, err)
			defer conn.Close()

			client := storage_backend.NewStorageBackendClient(conn)

			// when
			res, err := client.OnCreate(ctx, req)

			// no modification of additional params, asserting nil
			assert.Equal(t, testCase.ExpectedProviderState, testCase.InputProvider.secrets)

			// then
			if testCase.ExpectedErrorMessage != nil {
				assert.Nil(t, res)
				require.Error(t, err)
				assert.EqualError(t, err, *testCase.ExpectedErrorMessage)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Nil(t, res.AdditionalParameters)
		})
	}
}

func TestHandler_OnUpdate(t *testing.T) {
	// given
	providerName := "fake"
	reqAdditionalParams := []byte(fmt.Sprintf(`{"provider":"%s"}`, providerName))
	valueBytes := []byte(`{"key": true}`)
	req := &storage_backend.OnUpdateRequest{
		TypeinstanceId:       "uuid",
		NewResourceVersion:   3,
		NewValue:             valueBytes,
		AdditionalParameters: reqAdditionalParams,
	}

	path := fmt.Sprintf("/capact/%s", req.TypeinstanceId)

	testCases := []struct {
		Name                  string
		InputProvider         *fakeProvider
		ExpectedProviderState map[string]map[string]string
		ExpectedErrorMessage  *string
	}{
		{
			Name:          "No data", // data for a give nrevision could reside in different storage backend
			InputProvider: &fakeProvider{secrets: map[string]map[string]string{}},
			ExpectedProviderState: map[string]map[string]string{
				path: {
					"3": string(valueBytes),
				},
			},
		},
		{
			Name: "Empty value",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {},
				},
			},
			ExpectedProviderState: map[string]map[string]string{
				path: {
					"3": string(valueBytes),
				},
			},
		},
		{
			Name: "Already existing without conflict",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {
						"1":         "original",
						"locked_by": "service/foo",
					},
				},
			},
			ExpectedProviderState: map[string]map[string]string{
				path: {
					"1":         "original",
					"3":         string(valueBytes),
					"locked_by": "service/foo",
				},
			},
		},
		{
			Name: "Already existing with conflict",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {
						"3": "original",
					},
				},
			},
			ExpectedProviderState: map[string]map[string]string{
				path: {
					"3": "original",
				},
			},
			ExpectedErrorMessage: ptr.String("rpc error: code = AlreadyExists desc = entry \"/capact/uuid\" in provider \"fake\" already exist"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			srv, listener := setupServerAndListener(t, map[string]tellercore.Provider{
				providerName: testCase.InputProvider,
			})
			defer srv.Stop()

			ctx := context.Background()
			conn, err := grpc.DialContext(ctx, "", dialOpts(listener)...)
			require.NoError(t, err)
			defer conn.Close()

			client := storage_backend.NewStorageBackendClient(conn)

			// when
			res, err := client.OnUpdate(ctx, req)

			// no modification of additional params, asserting nil
			assert.Equal(t, testCase.ExpectedProviderState, testCase.InputProvider.secrets)

			// then
			if testCase.ExpectedErrorMessage != nil {
				assert.Nil(t, res)
				require.Error(t, err)
				assert.EqualError(t, err, *testCase.ExpectedErrorMessage)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Nil(t, res.AdditionalParameters)
		})
	}
}

func TestHandler_OnLock(t *testing.T) {
	// given
	providerName := "fake"
	reqAdditionalParams := []byte(fmt.Sprintf(`{"provider":"%s"}`, providerName))
	req := &storage_backend.OnLockRequest{
		TypeinstanceId:       "uuid",
		LockedBy:             "foo/sample",
		AdditionalParameters: reqAdditionalParams,
	}

	path := fmt.Sprintf("/capact/%s", req.TypeinstanceId)

	testCases := []struct {
		Name                  string
		InputProvider         *fakeProvider
		ExpectedProviderState map[string]map[string]string
		ExpectedErrorMessage  *string
	}{
		{
			Name:          "No data", // data for a give nrevision could reside in different storage backend
			InputProvider: &fakeProvider{secrets: map[string]map[string]string{}},
			ExpectedProviderState: map[string]map[string]string{
				path: {
					"locked_by": "foo/sample",
				},
			},
		},
		{
			Name: "Empty value",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {},
				},
			},
			ExpectedProviderState: map[string]map[string]string{
				path: {
					"locked_by": "foo/sample",
				},
			},
		},
		{
			Name: "Already existing without conflict",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {
						"1": "original",
					},
				},
			},
			ExpectedProviderState: map[string]map[string]string{
				path: {
					"1":         "original",
					"locked_by": "foo/sample",
				},
			},
		},
		{
			Name: "Already existing with conflict",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {
						"3":         "original",
						"locked_by": "previous",
					},
				},
			},
			ExpectedProviderState: map[string]map[string]string{
				path: {
					"3":         "original",
					"locked_by": "foo/sample",
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			srv, listener := setupServerAndListener(t, map[string]tellercore.Provider{
				providerName: testCase.InputProvider,
			})
			defer srv.Stop()

			ctx := context.Background()
			conn, err := grpc.DialContext(ctx, "", dialOpts(listener)...)
			require.NoError(t, err)
			defer conn.Close()

			client := storage_backend.NewStorageBackendClient(conn)

			// when
			res, err := client.OnLock(ctx, req)

			// no modification of additional params, asserting nil
			assert.Equal(t, testCase.ExpectedProviderState, testCase.InputProvider.secrets)

			// then
			if testCase.ExpectedErrorMessage != nil {
				assert.Nil(t, res)
				require.Error(t, err)
				assert.EqualError(t, err, *testCase.ExpectedErrorMessage)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, res)
		})
	}
}

func TestHandler_OnUnlock(t *testing.T) {
	// given
	providerName := "fake"
	reqAdditionalParams := []byte(fmt.Sprintf(`{"provider":"%s"}`, providerName))
	req := &storage_backend.OnUnlockRequest{
		TypeinstanceId:       "uuid",
		AdditionalParameters: reqAdditionalParams,
	}

	path := fmt.Sprintf("/capact/%s", req.TypeinstanceId)

	testCases := []struct {
		Name                  string
		InputProvider         *fakeProvider
		ExpectedProviderState map[string]map[string]string
		ExpectedErrorMessage  *string
	}{
		{
			Name:                  "No data", // data for a give nrevision could reside in different storage backend
			InputProvider:         &fakeProvider{secrets: map[string]map[string]string{}},
			ExpectedProviderState: map[string]map[string]string{},
		},
		{
			Name: "Already existing without conflict",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {
						"1":         "original",
						"locked_by": "foo/bar",
					},
				},
			},
			ExpectedProviderState: map[string]map[string]string{
				path: {
					"1": "original",
				},
			},
		},
		{
			Name: "Already existing empty property",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {
						"3":         "original",
						"locked_by": "",
					},
				},
			},
			ExpectedProviderState: map[string]map[string]string{
				path: {
					"3": "original",
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			srv, listener := setupServerAndListener(t, map[string]tellercore.Provider{
				providerName: testCase.InputProvider,
			})
			defer srv.Stop()

			ctx := context.Background()
			conn, err := grpc.DialContext(ctx, "", dialOpts(listener)...)
			require.NoError(t, err)
			defer conn.Close()

			client := storage_backend.NewStorageBackendClient(conn)

			// when
			res, err := client.OnUnlock(ctx, req)

			// no modification of additional params, asserting nil
			assert.Equal(t, testCase.ExpectedProviderState, testCase.InputProvider.secrets)

			// then
			if testCase.ExpectedErrorMessage != nil {
				assert.Nil(t, res)
				require.Error(t, err)
				assert.EqualError(t, err, *testCase.ExpectedErrorMessage)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, res)
		})
	}
}

func TestHandler_OnDelete(t *testing.T) {
	// given
	providerName := "fake"
	reqAdditionalParams := []byte(fmt.Sprintf(`{"provider":"%s"}`, providerName))
	req := &storage_backend.OnDeleteRequest{
		TypeinstanceId:       "uuid",
		AdditionalParameters: reqAdditionalParams,
	}

	path := fmt.Sprintf("/capact/%s", req.TypeinstanceId)

	testCases := []struct {
		Name                  string
		InputProvider         *fakeProvider
		ExpectedProviderState map[string]map[string]string
		ExpectedErrorMessage  *string
	}{
		{
			Name:                  "No data", // data for a give nrevision could reside in different storage backend
			InputProvider:         &fakeProvider{secrets: map[string]map[string]string{}},
			ExpectedProviderState: map[string]map[string]string{},
		},
		{
			Name: "Empty value",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {},
				},
			},
			ExpectedProviderState: map[string]map[string]string{},
		},
		{
			Name: "Already existing",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {
						"1":         "original",
						"locked_by": "foo/bar",
					},
					"cant-touch-this": {
						"Music":        "hits me so hard",
						"Makes me say": "Oh, my Lord",
					},
				},
			},
			ExpectedProviderState: map[string]map[string]string{
				"cant-touch-this": {
					"Music":        "hits me so hard",
					"Makes me say": "Oh, my Lord",
				},
			},
		},
		{
			Name: "Other data",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					"cant-touch-this": {
						"Music":        "hits me so hard",
						"Makes me say": "Oh, my Lord",
					},
				},
			},
			ExpectedProviderState: map[string]map[string]string{
				"cant-touch-this": {
					"Music":        "hits me so hard",
					"Makes me say": "Oh, my Lord",
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			srv, listener := setupServerAndListener(t, map[string]tellercore.Provider{
				providerName: testCase.InputProvider,
			})
			defer srv.Stop()

			ctx := context.Background()
			conn, err := grpc.DialContext(ctx, "", dialOpts(listener)...)
			require.NoError(t, err)
			defer conn.Close()

			client := storage_backend.NewStorageBackendClient(conn)

			// when
			res, err := client.OnDelete(ctx, req)

			// no modification of additional params, asserting nil
			assert.Equal(t, testCase.ExpectedProviderState, testCase.InputProvider.secrets)

			// then
			if testCase.ExpectedErrorMessage != nil {
				assert.Nil(t, res)
				require.Error(t, err)
				assert.EqualError(t, err, *testCase.ExpectedErrorMessage)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, res)
		})
	}
}

const bufSize = 1024 * 1024

func setupServerAndListener(t *testing.T, providersMap map[string]tellercore.Provider) (*grpc.Server, *bufconn.Listener) {
	t.Helper()
	handler := secret_storage_backend.NewHandler(logger.Noop(), providersMap)

	listener := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	storage_backend.RegisterStorageBackendServer(srv, handler)

	go func() {
		err := srv.Serve(listener)
		require.NoError(t, err)
	}()

	return srv, listener
}

func dialOpts(listener *bufconn.Listener) []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithInsecure(),
	}
}
