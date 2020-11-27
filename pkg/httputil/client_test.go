package httputil_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"projectvoltron.dev/voltron/pkg/httputil"
)

func TestNewClient(t *testing.T) {
	tests := map[string]struct {
		timeout              time.Duration
		skipCertVerification bool
	}{
		"Respect given timeout and skips cert verification": {
			timeout:              time.Second,
			skipCertVerification: true,
		},
		"Respect given timeout and DOES NOT skip cert verification": {
			timeout:              time.Second,
			skipCertVerification: false,
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// when
			gotCli := httputil.NewClient(tc.timeout, tc.skipCertVerification)

			// then
			transport := gotCli.Transport.(*http.Transport)
			assert.Equal(t, tc.skipCertVerification, transport.TLSClientConfig.InsecureSkipVerify)
			assert.Equal(t, gotCli.Timeout, tc.timeout)

			// assert that at least some default timeouts are set
			assert.NotZero(t, transport.IdleConnTimeout)
			assert.NotZero(t, transport.TLSHandshakeTimeout)
			assert.NotZero(t, transport.ExpectContinueTimeout)
		})
	}
}

func TestClientBasicAuth(t *testing.T) {
	tt := map[string]struct {
		username            string
		password            string
		exptectedStatusCode int
	}{
		"Should use basic auth when making request": {
			username:            "test",
			password:            "s3cr3t",
			exptectedStatusCode: 200,
		},
	}

	for tn, tc := range tt {
		t.Run(tn, func(t *testing.T) {
			cli := httputil.NewClient(30*time.Second, false, httputil.WithBasicAuth(tc.username, tc.password))

			url := fmt.Sprintf("https://httpbin.org/basic-auth/%s/%s", tc.username, tc.password)
			resp, err := cli.Get(url)

			require.Nil(t, err, "error making http request: %v", err)
			defer resp.Body.Close()
			require.Equal(t, tc.exptectedStatusCode, resp.StatusCode, "incorrect status code")
		})
	}
}
