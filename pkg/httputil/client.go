package httputil

import (
	"net/http"
	"time"
)

const defaultTimeout = 30 * time.Second

// NewClient creates a new http client with a given timeouts
func NewClient(opts ...ClientOption) *http.Client {
	client := &http.Client{
		Transport: newConfigurableTransport(),
		Timeout:   defaultTimeout,
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

// WithTimeout returns a ClientOption to set a given timeout for client.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(client *http.Client) {
		client.Timeout = timeout
	}
}

// WithTLSInsecureSkipVerify returns a ClientOption to skip TLS verification for the HTTP server.
func WithTLSInsecureSkipVerify(skip bool) func(client *http.Client) {
	return func(client *http.Client) {
		client.Transport.(*configurableTransport).SetTLSInsecureSkipVerify(skip)
	}
}
