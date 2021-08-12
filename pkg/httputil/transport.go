package httputil

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

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

func (t *configurableTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if t.user != "" && t.pass != "" {
		req.SetBasicAuth(t.user, t.pass)
	}
	now := time.Now()
	defer func() {
		fmt.Printf("RoundTrip time for %v: %v", req.URL.String(), time.Since(now))
		fmt.Printf("Request size: %v", strconv.FormatInt(req.ContentLength, 10))
		fmt.Printf("Response size: %v", strconv.FormatInt(resp.ContentLength, 10))
		//b, err = io.Copy(ioutil.Discard, resp.ContentLength)
	}()
	return t.transport.RoundTrip(req)
}
