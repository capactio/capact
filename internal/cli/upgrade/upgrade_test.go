package upgrade

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"capact.io/capact/pkg/httputil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetLatestVersion tests that the latest version of Helm chart is returned.
// The httptest.Server was not used as the logic is based on the URL, so the HTTP client has to be mocked to validate that:
//
// - for master URLs Helm charts the latest entry is selected base on the `Created` field,
// - for all other URLs the latest entry is selected based on SemVer.
//
func TestGetLatestVersion(t *testing.T) {
	tests := []struct {
		name            string
		url             string
		expectedVersion string
	}{
		{
			name:            "Master URL should sort by Created timestamp",
			url:             capactioHelmRepoMaster,
			expectedVersion: "0.2.0-7a347a9",
		},
		{
			name:            "Master URL should sort by Created timestamp",
			url:             CapactioHelmRepoOfficial,
			expectedVersion: "0.2.1",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			fakeCli := httputil.NewClient(0)
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
			upgrade := Upgrade{
				httpClient: fakeCli,
			}

			// when
			ver, err := upgrade.getLatestVersion(tc.url)

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
