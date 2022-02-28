package secretstoragebackend_test

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
	reqContext := []byte(fmt.Sprintf(`{"provider":"%s"}`, providerName))
	req := &storage_backend.GetValueRequest{
		TypeinstanceId:  "uuid",
		ResourceVersion: 2,
		Context:         reqContext,
	}
	path := fmt.Sprintf("/capact/%s", req.TypeinstanceId)

	testCases := []struct {
		Name                 string
		InputProvider        tellercore.Provider
		ExpectedValue        []byte
		ExpectedErrorMessage *string
	}{
		{
			Name:                 "No secret",
			InputProvider:        &fakeProvider{},
			ExpectedValue:        nil,
			ExpectedErrorMessage: ptr.String("rpc error: code = NotFound desc = TypeInstance \"uuid\" in revision 2 was not found"),
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
			if testCase.ExpectedErrorMessage != nil {
				assert.Nil(t, res)
				require.Error(t, err)
				assert.EqualError(t, err, *testCase.ExpectedErrorMessage)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, testCase.ExpectedValue, res.Value)
		})
	}
}

func TestHandler_GetLockedBy(t *testing.T) {
	// given
	providerName := "fake"
	reqContext := []byte(fmt.Sprintf(`{"provider":"%s"}`, providerName))
	req := &storage_backend.GetLockedByRequest{
		TypeinstanceId: "uuid",
		Context:        reqContext,
	}
	path := fmt.Sprintf("/capact/%s", req.TypeinstanceId)

	testCases := []struct {
		Name                 string
		InputProvider        tellercore.Provider
		ExpectedLockedBy     *string
		ExpectedErrorMessage *string
	}{
		{
			Name:                 "No data",
			InputProvider:        &fakeProvider{},
			ExpectedLockedBy:     nil,
			ExpectedErrorMessage: ptr.String("rpc error: code = NotFound desc = TypeInstance \"uuid\" not found: secret from path \"/capact/uuid\" is empty"),
		},
		{
			Name: "Empty value",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {
						"1": "bar",
					},
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
			if testCase.ExpectedErrorMessage != nil {
				assert.Nil(t, res)
				require.Error(t, err)
				assert.EqualError(t, err, *testCase.ExpectedErrorMessage)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Equal(t, testCase.ExpectedLockedBy, res.LockedBy)
		})
	}
}

func TestHandler_OnCreate(t *testing.T) {
	// given
	providerName := "fake"
	reqContext := []byte(fmt.Sprintf(`{"provider":"%s"}`, providerName))
	valueBytes := []byte(`{"key": true}`)
	req := &storage_backend.OnCreateRequest{
		TypeinstanceId: "uuid",
		Value:          valueBytes,
		Context:        reqContext,
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
			Name: "Already existing",
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
			ExpectedErrorMessage: ptr.String("rpc error: code = AlreadyExists desc = path \"/capact/uuid\" in provider \"fake\" already exist"),
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
			assert.Nil(t, res.Context)
		})
	}
}

func TestHandler_OnUpdate(t *testing.T) {
	// given
	providerName := "fake"
	reqContext := []byte(fmt.Sprintf(`{"provider":"%s"}`, providerName))
	valueBytes := []byte(`{"key": true}`)
	req := &storage_backend.OnUpdateRequest{
		TypeinstanceId:     "uuid",
		NewResourceVersion: 3,
		NewValue:           valueBytes,
		Context:            reqContext,
	}

	path := fmt.Sprintf("/capact/%s", req.TypeinstanceId)

	testCases := []struct {
		Name                  string
		InputProvider         *fakeProvider
		ExpectedProviderState map[string]map[string]string
		ExpectedErrorMessage  *string
	}{
		{
			Name: "Non-existing secret",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {},
				},
			},
			ExpectedProviderState: map[string]map[string]string{
				path: {},
			},
			ExpectedErrorMessage: ptr.String("rpc error: code = NotFound desc = path \"/capact/uuid\" in provider \"fake\" not found"),
		},
		{
			Name: "Already existing locked",
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
					"locked_by": "service/foo",
				},
			},
			ExpectedErrorMessage: ptr.String("rpc error: code = FailedPrecondition desc = typeInstance locked: path \"/capact/uuid\" contains \"locked_by\" property with value \"service/foo\""),
		},
		{
			Name: "Already existing not locked",
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
			ExpectedErrorMessage: ptr.String("rpc error: code = AlreadyExists desc = field \"3\" for path \"/capact/uuid\" in provider \"fake\" already exist"),
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
			assert.Nil(t, res.Context)
		})
	}
}

func TestHandler_OnLock(t *testing.T) {
	// given
	providerName := "fake"
	reqContext := []byte(fmt.Sprintf(`{"provider":"%s"}`, providerName))
	req := &storage_backend.OnLockRequest{
		TypeinstanceId: "uuid",
		LockedBy:       "foo/sample",
		Context:        reqContext,
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
			Name: "Already existing locked",
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
					"locked_by": "previous",
				},
			},
			ExpectedErrorMessage: ptr.String("rpc error: code = FailedPrecondition desc = typeInstance locked: path \"/capact/uuid\" contains \"locked_by\" property with value \"previous\""),
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
	reqContext := []byte(fmt.Sprintf(`{"provider":"%s"}`, providerName))
	req := &storage_backend.OnUnlockRequest{
		TypeinstanceId: "uuid",
		Context:        reqContext,
	}

	path := fmt.Sprintf("/capact/%s", req.TypeinstanceId)

	testCases := []struct {
		Name                  string
		InputProvider         *fakeProvider
		ExpectedProviderState map[string]map[string]string
		ExpectedErrorMessage  *string
	}{
		{
			Name:                  "No data",
			InputProvider:         &fakeProvider{secrets: map[string]map[string]string{}},
			ExpectedProviderState: map[string]map[string]string{},
			ExpectedErrorMessage:  ptr.String("rpc error: code = NotFound desc = path \"/capact/uuid\" in provider \"fake\" not found"),
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
		{
			Name: "Already existing without lockedBy property",
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
	reqContext := []byte(fmt.Sprintf(`{"provider":"%s"}`, providerName))
	req := &storage_backend.OnDeleteRequest{
		TypeinstanceId: "uuid",
		Context:        reqContext,
	}

	path := fmt.Sprintf("/capact/%s", req.TypeinstanceId)

	testCases := []struct {
		Name                  string
		InputProvider         *fakeProvider
		ExpectedProviderState map[string]map[string]string
		ExpectedErrorMessage  *string
	}{
		{
			Name:                  "No data",
			InputProvider:         &fakeProvider{secrets: map[string]map[string]string{}},
			ExpectedProviderState: map[string]map[string]string{},
			ExpectedErrorMessage:  ptr.String("rpc error: code = NotFound desc = path \"/capact/uuid\" in provider \"fake\" not found"),
		},
		{
			Name: "Empty value",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {},
				},
			},
			ExpectedProviderState: map[string]map[string]string{path: {}},
			ExpectedErrorMessage:  ptr.String("rpc error: code = NotFound desc = path \"/capact/uuid\" in provider \"fake\" not found"),
		},
		{
			Name: "Already existing not locked",
			InputProvider: &fakeProvider{
				secrets: map[string]map[string]string{
					path: {
						"1": "original",
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
			Name: "Already existing locked",
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
					"1":         "original",
					"locked_by": "foo/bar",
				},
			},
			ExpectedErrorMessage: ptr.String("rpc error: code = FailedPrecondition desc = typeInstance locked: path \"/capact/uuid\" contains \"locked_by\" property with value \"foo/bar\""),
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
