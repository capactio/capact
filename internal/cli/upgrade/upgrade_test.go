package upgrade

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"capact.io/capact/internal/cli/capact"
	"capact.io/capact/pkg/httputil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetLatestVersion tests that the latest version of Helm chart is returned.
// The httptest.Server was not used as the logic is based on the URL, so the HTTP client has to be mocked to validate that:
//
// - for the @latest URL Helm charts the latest entry is selected base on the `Created` field,
// - for all other URLs the latest entry is selected based on SemVer.
//
func TestGetLatestVersion(t *testing.T) {
	tests := []struct {
		name            string
		url             string
		expectedVersion string
	}{
		{
			name:            "Latest URL should sort by Created timestamp",
			url:             capact.HelmRepoLatest,
			expectedVersion: "0.2.0-7a347a9",
		},
		{
			name:            "Official URL should sort by version",
			url:             capact.HelmRepoStable,
			expectedVersion: "0.2.1",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			fakeCli := httputil.NewClient()
			// mock transport to do not execute a real call
			fakeCli.Transport = roundTripperFunc(func(request *http.Request) (*http.Response, error) {
				assert.True(t, strings.HasPrefix(request.URL.String(), tc.url))

				file, err := os.Open("testdata/TestGetLatestVersion/index.yaml")
				if err != nil {
					return nil, err
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       file,
				}, nil
			})
			helper := capact.HelmHelper{
				HTTPClient: fakeCli,
			}

			// when
			ver, err := helper.GetLatestVersion(tc.url, "capact")

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.expectedVersion, ver)
		})
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
