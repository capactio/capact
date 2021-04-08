package httputil

import (
	"net/http"
	"time"
)

// NewClient creates a new http client with a given timeouts
func NewClient(timeout time.Duration, opts ...ClientOption) *http.Client {
	client := &http.Client{
		Transport: http.DefaultTransport.(*http.Transport).Clone(),
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

func WithTLSInsecureSkipVerify(skip bool) func(client *http.Client) {
	return func(client *http.Client) {
		tr := client.Transport.(*http.Transport)
		tr.TLSClientConfig.InsecureSkipVerify = skip
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
