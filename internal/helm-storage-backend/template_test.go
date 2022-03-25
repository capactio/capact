package helmstoragebackend

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/time"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/logger"
	"capact.io/capact/internal/ptr"
	pb "capact.io/capact/pkg/hub/api/grpc/storage_backend"
)

func TestTemplate_CreateGetAndUpdate_Success(t *testing.T) {
	tests := []struct {
		name string

		givenDriver          *string
		givenTypeInstanceID  string
		givenResourceVersion uint32
		expectedDriver       string
	}{
		{
			name:                "should use default driver and return the latest release",
			givenTypeInstanceID: "123",
			givenDriver:         nil,
			expectedDriver:      "secrets",
		},
		{
			name:                "should use configmap driver and return the latest release",
			givenTypeInstanceID: "123",
			givenDriver:         ptr.String("configmaps"),
			expectedDriver:      "configmaps",
		},
		{
			name:                 "should ignore resourceVersion and return the latest release",
			givenTypeInstanceID:  "123",
			givenResourceVersion: 42, // should be ignored
			expectedDriver:       "secrets",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// given
			const (
				releaseName      = "test-get-release"
				releaseNamespace = "test-get-namespace"
			)

			schart, err := loader.Load("./testdata/sample-chart")
			require.NoError(t, err)

			expHelmRelease := fixHelmReleaseWithChart(releaseName, releaseNamespace, schart)
			expFlags := &genericclioptions.ConfigFlags{ClusterName: ptr.String("testing")}
			mockConfigurationProducer := mockConfigurationProducer(t, expHelmRelease, expFlags, test.expectedDriver)

			givenReq := &pb.GetValueRequest{
				TypeInstanceId:  test.givenTypeInstanceID,
				ResourceVersion: test.givenResourceVersion,
				Context: mustMarshal(t, TemplateContext{
					HelmRelease: HelmRelease{
						Name:      releaseName,
						Namespace: releaseNamespace,
						Driver:    test.givenDriver,
					},
					GoTemplate: heredoc.Doc(`
							host: '{{ include "sample-chart.fullname" . }}'
							port: '{{ .Values.service.port }}'
							superuser:
							  username: 'psql'
							`),
				})}

			expResponse := &pb.GetValueResponse{
				Value: mustMarshal(t, map[string]interface{}{
					"host": "test-get-release-sample-chart",
					"port": "80",
					"superuser": map[string]interface{}{
						"username": "psql",
					},
				}),
			}

			fetcher := NewHelmReleaseFetcher(expFlags)
			fetcher.actionConfigurationProducer = mockConfigurationProducer
			svc := NewTemplateHandler(logger.Noop(), fetcher)

			// when
			outVal, gotErr := svc.GetValue(context.Background(), givenReq)

			// then
			assert.NoError(t, gotErr)
			assert.EqualValues(t, outVal, expResponse)

			// when
			createOut, createErr := svc.OnCreate(context.Background(), &pb.OnCreateRequest{
				TypeInstanceId: givenReq.TypeInstanceId,
				Context:        givenReq.Context,
			})

			// then
			assert.NoError(t, createErr)
			assert.Empty(t, createOut)

			// when
			updateOut, updateErr := svc.OnUpdate(context.Background(), &pb.OnUpdateRequest{
				TypeInstanceId: givenReq.TypeInstanceId,
				Context:        givenReq.Context,
			})

			// then
			assert.NoError(t, updateErr)
			assert.Empty(t, updateOut)
		})
	}
}

func TestTemplate_CreateGetAndUpdate_Failures(t *testing.T) {
	// globally given
	const (
		releaseName      = "test-release"
		releaseNamespace = "test-namespace"
	)
	tests := []struct {
		name string

		request       *pb.GetValueRequest
		internalError error

		expErrMsg string
	}{
		{
			name: "should return not found error if release name is wrong",
			request: &pb.GetValueRequest{
				TypeInstanceId: "123",
				Context: mustMarshal(t, TemplateContext{
					HelmRelease: HelmRelease{
						Name:      "other-release",
						Namespace: releaseNamespace,
					},
				}),
			},
			expErrMsg: "rpc error: code = NotFound desc = Helm release 'test-namespace/other-release' for TypeInstance '123' was not found",
		},
		{
			name: "should return not found error if release namespace is wrong",
			request: &pb.GetValueRequest{
				TypeInstanceId: "123",
				Context: mustMarshal(t, TemplateContext{
					HelmRelease: HelmRelease{
						Name:      releaseName,
						Namespace: "other-ns",
					},
				}),
			},
			expErrMsg: "rpc error: code = NotFound desc = Helm release 'other-ns/test-release' for TypeInstance '123' was not found",
		},
		{
			name: "should return error indicating invalid goTemplate",
			request: &pb.GetValueRequest{
				TypeInstanceId: "123",
				Context: mustMarshal(t, TemplateContext{
					HelmRelease: HelmRelease{
						Name:      releaseName,
						Namespace: releaseNamespace,
					},
					GoTemplate: `host: '{{ .Missing.property }}'`,
				}),
			},
			expErrMsg: "rpc error: code = Internal desc = while rendering output value: while rendering additional output: while rendering chart: template: test-release-chart/additionalOutputTemplate:1:18: executing \"test-release-chart/additionalOutputTemplate\" at <.Missing.property>: nil pointer evaluating interface {}.property",
		},
		{
			name: "should return internal error",
			request: &pb.GetValueRequest{
				TypeInstanceId: "123",
				Context: mustMarshal(t, TemplateContext{
					HelmRelease: HelmRelease{
						Name:      releaseName,
						Namespace: "other-ns",
					},
				}),
			},
			internalError: errors.New("internal error"),
			expErrMsg:     "rpc error: code = Internal desc = while creating Helm get release client: internal error",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// given
			expHelmRelease := fixHelmRelease(releaseName, releaseNamespace)
			expFlags := &genericclioptions.ConfigFlags{ClusterName: ptr.String("testing")}

			mockConfigurationProducer := func(inputFlags *genericclioptions.ConfigFlags, inputDriver, inputNs string) (*action.Configuration, error) {
				if test.internalError != nil {
					return nil, test.internalError
				}
				producer := mockConfigurationProducer(t, expHelmRelease, expFlags, "secrets")
				return producer(inputFlags, inputDriver, inputNs)
			}
			fetcher := NewHelmReleaseFetcher(expFlags)
			fetcher.actionConfigurationProducer = mockConfigurationProducer

			svc := NewTemplateHandler(logger.Noop(), fetcher)

			// when
			outVal, gotErr := svc.GetValue(context.Background(), test.request)

			// then
			assert.EqualError(t, gotErr, test.expErrMsg)
			assert.Nil(t, outVal)

			// when
			createOut, createErr := svc.OnCreate(context.Background(), &pb.OnCreateRequest{
				TypeInstanceId: test.request.TypeInstanceId,
				Context:        test.request.Context,
			})

			// then
			assert.EqualError(t, createErr, test.expErrMsg)
			assert.Empty(t, createOut)

			// when
			updateOut, updateErr := svc.OnUpdate(context.Background(), &pb.OnUpdateRequest{
				TypeInstanceId: test.request.TypeInstanceId,
				Context:        test.request.Context,
			})

			// then
			assert.EqualError(t, updateErr, test.expErrMsg)
			assert.Empty(t, updateOut)
		})
	}
}

func TestTemplate_NOP_Methods(t *testing.T) {
	// globally given
	tests := []struct {
		name    string
		handler func(ctx context.Context, svc *TemplateHandler) (interface{}, error)
	}{
		{
			name: "no operation for OnDelete",
			handler: func(ctx context.Context, svc *TemplateHandler) (interface{}, error) {
				return svc.OnDelete(ctx, nil)
			},
		},
		{
			name: "no operation for GetLockedBy",
			handler: func(ctx context.Context, svc *TemplateHandler) (interface{}, error) {
				return svc.GetLockedBy(ctx, nil)
			},
		},
		{
			name: "no operation for OnLock",
			handler: func(ctx context.Context, svc *TemplateHandler) (interface{}, error) {
				return svc.OnLock(ctx, nil)
			},
		},
		{
			name: "no operation for OnUnlock",
			handler: func(ctx context.Context, svc *TemplateHandler) (interface{}, error) {
				return svc.OnUnlock(ctx, nil)
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// given
			producerCalled := false
			mockConfigurationProducer := func(_ *genericclioptions.ConfigFlags, _, _ string) (*action.Configuration, error) {
				producerCalled = true
				return nil, nil
			}
			fetcher := NewHelmReleaseFetcher(nil)
			fetcher.actionConfigurationProducer = mockConfigurationProducer
			svc := NewTemplateHandler(logger.Noop(), fetcher)

			// when
			outVal, gotErr := test.handler(context.Background(), svc)

			// then
			assert.NoError(t, gotErr)
			assert.False(t, producerCalled)
			assert.Empty(t, outVal)
		})
	}
}

func fixHelmReleaseWithChart(name, ns string, chrt *chart.Chart) *release.Release {
	now := time.Now()
	return &release.Release{
		Name:      name,
		Namespace: ns,
		Info: &release.Info{
			FirstDeployed: now,
			LastDeployed:  now,
			Description:   "Named Release Stub",
		},
		Chart: chrt,
	}
}
