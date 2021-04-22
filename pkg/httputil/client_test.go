package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := map[string]struct {
		timeout              time.Duration
		skipCertVerification bool
		username             string
		password             string
	}{
		"Set given timeout and skips cert verification": {
			timeout:              time.Second,
			skipCertVerification: true,
		},
		"Set given timeout and DOES NOT skip cert verification": {
			timeout:              time.Second,
			skipCertVerification: false,
		},
		"Set given basic auth": {
			username: "abc",
			password: "abc",
		},
		"Set given basic auth and skip cert verification": {
			username:             "abc",
			password:             "abc",
			skipCertVerification: true,
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			var receivedUser, receivedPass string
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedUser, receivedPass, _ = r.BasicAuth()
			}))

			cli := NewClient(tc.timeout, WithBasicAuth(tc.username, tc.password), WithTLSInsecureSkipVerify(tc.skipCertVerification))

			// when
			resp, err := cli.Get(ts.URL)

			// then
			require.NoError(t, err)
			defer resp.Body.Close()

			// BasicAuth is configured
			require.Equal(t, tc.username, receivedUser)
			require.Equal(t, tc.password, receivedPass)

			// TLS is configured
			cfgTransport := cli.Transport.(*ConfigurableTransport)
			assert.Equal(t, tc.skipCertVerification, cfgTransport.transport.TLSClientConfig.InsecureSkipVerify)
			assert.Equal(t, tc.timeout, cli.Timeout)

			// assert that at least some default timeouts are set
			assert.NotZero(t, cfgTransport.transport.IdleConnTimeout)
			assert.NotZero(t, cfgTransport.transport.TLSHandshakeTimeout)
			assert.NotZero(t, cfgTransport.transport.ExpectContinueTimeout)
		})
	}
}
