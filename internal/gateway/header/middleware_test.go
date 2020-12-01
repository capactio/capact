package header_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"projectvoltron.dev/voltron/internal/gateway/header"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestSaveHeadersInCtxHTTPMiddleware(t *testing.T) {
	// given
	req := sampleRequest()
	req.Header.Set("foo", "bar")
	req.Header.Set("baz", "qux")

	expectedHeaders := req.Header.Clone()

	testHandler := setupHandler(expectedHeaders)

	rw := httptest.NewRecorder()

	handler := header.SaveHeadersInCtxHTTPMiddleware(testHandler)

	// when
	handler.ServeHTTP(rw, req)

	// then
	assert.Equal(t, http.StatusOK, rw.Code)
}

func TestSetHeadersFromCtxGatewayMiddleware(t *testing.T) {
	// given
	req := sampleRequest()
	req.Header["do-not-overwrite"] = []string{"not-overwritten"}

	var headers http.Header = map[string][]string{
		"foo":              {"bar", "baz"},
		"bar":              {"baz"},
		"baz":              {"qux"},
		"do-not-overwrite": {"overwritten"},
	}

	var expectedHeaders http.Header = map[string][]string{
		"foo":              {"bar", "baz"},
		"bar":              {"baz"},
		"baz":              {"qux"},
		"do-not-overwrite": {"not-overwritten"},
	}

	ctxWithHeaders := header.NewContext(context.Background(), headers)
	middlewareFn := header.SetHeadersFromCtxGQLMiddleware()

	reqWithCtx := req.WithContext(ctxWithHeaders)

	// when
	err := middlewareFn(reqWithCtx)

	// then
	require.NoError(t, err)
	assert.Equal(t, expectedHeaders, reqWithCtx.Header)
}

func sampleRequest() *http.Request {
	return httptest.NewRequest("GET", "/", strings.NewReader(""))
}

func setupHandler(expectedValue http.Header) http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		actualValue, ok := header.FromContext(req.Context())
		if !ok {
			http.Error(w, "no headers in context", http.StatusInternalServerError)
			return
		}

		if !reflect.DeepEqual(actualValue, expectedValue) {
			http.Error(w, fmt.Sprintf("different headers in context: actual: %s; expected: %s", actualValue, expectedValue), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	return router
}
