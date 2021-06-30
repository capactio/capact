package httputil

import (
	"net/http"
	"time"
)

// NewClient creates a new http client with a given timeouts
func NewClient(timeout time.Duration, opts ...ClientOption) *http.Client {
	client := &http.Client{
		Transport: newConfigurableTransport(),
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

// WithBasicAuth returns a ClientOption to add basic access authentication credentials.
func WithBasicAuth(user, pass string) ClientOption {
	return func(client *http.Client) {
		client.Transport.(*configurableTransport).SetBasicAuth(user, pass)
	}
}

// WithTLSInsecureSkipVerify returns a ClientOption to skip TLS verification for the HTTP server.
func WithTLSInsecureSkipVerify(skip bool) func(client *http.Client) {
	return func(client *http.Client) {
		client.Transport.(*configurableTransport).SetTLSInsecureSkipVerify(skip)
	}
}
