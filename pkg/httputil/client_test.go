package httputil_test

import (
	"net/http"
	"net/http/httptest"
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
			gotCli := httputil.NewClient(tc.timeout, httputil.WithTLSInsecureSkipVerify(tc.skipCertVerification))

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

func TestClientProvidedBasicAuthIsUsedInRequests(t *testing.T) {
	// given
	const (
		username = "test"
		password = "s3cr3t"
	)

	var receivedUser, receivedPass string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUser, receivedPass, _ = r.BasicAuth()
	}))

	cli := httputil.NewClient(30*time.Second, httputil.WithBasicAuth(username, password))

	// when
	resp, err := cli.Get(ts.URL)

	// then
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, username, receivedUser)
	require.Equal(t, password, receivedPass)
}
