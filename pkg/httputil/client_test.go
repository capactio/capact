package httputil_test

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"projectvoltron.dev/voltron/pkg/httputil"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := map[string]struct {
		timeout time.Duration
		skipCertVerification bool
	}{
		"Respect given timeout and skips cert verification": {
			timeout: time.Second,
			skipCertVerification: true,
		},
		"Respect given timeout and DOES NOT skip cert verification": {
			timeout: time.Second,
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
