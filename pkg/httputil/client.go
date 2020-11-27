package httputil

import (
	"net/http"
	"time"
)

// NewClient creates a new http client with a given timeouts
func NewClient(timeout time.Duration, skipCertVerification bool, opts ...ClientOption) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig.InsecureSkipVerify = skipCertVerification

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	for _, optionFunc := range opts {
		optionFunc(client)
	}

	return client
}

// ClientOption are functions that are passed into NewClient to
// modify the behaviour of the Client.
type ClientOption func(*http.Client)

func WithBasicAuth(user, pass string) ClientOption {
	return func(client *http.Client) {
		client.Transport = &authRoundTripper{
			user:         user,
			pass:         pass,
			RoundTripper: client.Transport,
		}
	}
}

type authRoundTripper struct {
	user string
	pass string
	http.RoundTripper
}

func (t *authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(t.user, t.pass)
	return t.RoundTripper.RoundTrip(req)
}
