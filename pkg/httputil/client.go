package httputil

import (
	"net/http"
	"time"
)

// NewClient creates a new http client with a given timeouts
func NewClient(timeout time.Duration, skipCertVerification bool) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig.InsecureSkipVerify = skipCertVerification

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}
