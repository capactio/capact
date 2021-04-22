package httputil

import "net/http"

// newConfigurableTransport
func newConfigurableTransport() *ConfigurableTransport {
	return &ConfigurableTransport{
		transport: http.DefaultTransport.(*http.Transport).Clone(),
	}
}

type ConfigurableTransport struct {
	user      string
	pass      string
	transport *http.Transport
}

func (t *ConfigurableTransport) SetBasicAuth(user, pass string) {
	t.user = user
	t.pass = pass
}

func (t *ConfigurableTransport) SetTLSInsecureSkipVerify(skip bool) {
	t.transport.TLSClientConfig.InsecureSkipVerify = skip
}

func (t *ConfigurableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(t.user, t.pass)
	return t.transport.RoundTrip(req)
}
