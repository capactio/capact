package capact

import (
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
)

func TestCreateObjectRetrySuccess(t *testing.T) {
	tests := map[string]struct {
		givenError error
	}{
		"Should pass on nil error": {
			givenError: nil,
		},
		"Should ignore AlreadyExist error": {
			givenError: apierrors.NewAlreadyExists(schema.GroupResource{}, "test-object"),
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			failingCli := &FailingKubeClient{
				CreateError: tc.givenError,
			}

			actionCfg := &action.Configuration{
				KubeClient: failingCli,
			}

			objToCreate := fmt.Sprintf(issuerTemplate, clusterIssuerName, certManagerSecretName)

			// when
			err := createObject(actionCfg, []byte(objToCreate))

			// then
			assert.NoError(t, err)
			assert.Equal(t, 1, failingCli.CreateCallsCnt)
		})
	}
}

func TestCreateObjectRetryFailed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestCreateObjectRetry as it takes up to 1.5s")
	}

	// given
	expTotalRetryDurationInSec := 1.6

	internalError := apierrors.NewInternalError(errors.New("simulate internal error"))
	failingCli := &FailingKubeClient{
		CreateError: internalError,
	}

	actionCfg := &action.Configuration{
		KubeClient: failingCli,
	}

	objToCreate := fmt.Sprintf(issuerTemplate, clusterIssuerName, certManagerSecretName)

	// when
	startedCall := time.Now()
	err := createObject(actionCfg, []byte(objToCreate))

	// then
	assert.Contains(t, err.Error(), internalError.Error())
	assert.Equal(t, 5, failingCli.CreateCallsCnt)
	assert.InDelta(t, expTotalRetryDurationInSec, time.Since(startedCall).Seconds(), 0.2)
}

// FailingKubeClient implements KubeClient for testing purposes.
type FailingKubeClient struct {
	CreateError    error
	CreateCallsCnt int
}

// Create returns the configured error.
func (f *FailingKubeClient) Create(resources kube.ResourceList) (*kube.Result, error) {
	f.CreateCallsCnt++
	return nil, f.CreateError
}

func (f *FailingKubeClient) Build(r io.Reader, _ bool) (kube.ResourceList, error) {
	return []*resource.Info{}, nil
}

func (f *FailingKubeClient) Wait(resources kube.ResourceList, d time.Duration) error {
	return errors.New("not implemented")
}

func (f *FailingKubeClient) WaitWithJobs(resources kube.ResourceList, d time.Duration) error {
	return errors.New("not implemented")
}

func (f *FailingKubeClient) Delete(resources kube.ResourceList) (*kube.Result, []error) {
	return nil, []error{errors.New("not implemented")}
}

func (f *FailingKubeClient) WatchUntilReady(resources kube.ResourceList, d time.Duration) error {
	return errors.New("not implemented")
}

func (f *FailingKubeClient) Update(r, modified kube.ResourceList, ignoreMe bool) (*kube.Result, error) {
	return nil, errors.New("not implemented")
}

func (f *FailingKubeClient) WaitAndGetCompletedPodPhase(s string, d time.Duration) (v1.PodPhase, error) {
	return v1.PodFailed, errors.New("not implemented")
}

func (f *FailingKubeClient) IsReachable() error {
	return nil
}
