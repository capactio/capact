package httputil

import "net/http"

// newConfigurableTransport
func newConfigurableTransport() *configurableTransport {
	return &configurableTransport{
		transport: http.DefaultTransport.(*http.Transport).Clone(),
	}
}

type configurableTransport struct {
	user      string
	pass      string
	transport *http.Transport
}

func (t *configurableTransport) SetBasicAuth(user, pass string) {
	t.user = user
	t.pass = pass
}

func (t *configurableTransport) SetTLSInsecureSkipVerify(skip bool) {
	t.transport.TLSClientConfig.InsecureSkipVerify = skip
}

func (t *configurableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(t.user, t.pass)
	return t.transport.RoundTrip(req)
}
