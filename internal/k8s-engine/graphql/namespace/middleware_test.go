package namespace_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"capact.io/capact/internal/k8s-engine/graphql/namespace"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware_Handle(t *testing.T) {
	// given
	customNS := "foo"

	reqWithCustomNS := sampleRequest()
	reqWithCustomNS.Header.Set(namespace.NamespaceHeaderName, customNS)

	tests := []struct {
		name              string
		inputRequest      *http.Request
		expectedNamespace string
	}{
		{
			inputRequest:      reqWithCustomNS,
			expectedNamespace: customNS,
		},
		{
			inputRequest:      sampleRequest(),
			expectedNamespace: namespace.DefaultNamespace,
		},
	}
	//nolint:scopelint
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			middleware := namespace.NewMiddleware()

			handler := setupHandler(middleware.Handle, testCase.expectedNamespace)

			rw := httptest.NewRecorder()

			// when
			handler.ServeHTTP(rw, testCase.inputRequest)

			// then
			assert.Equal(t, http.StatusOK, rw.Code)
		})
	}
}

func sampleRequest() *http.Request {
	return httptest.NewRequest("GET", "/", strings.NewReader(""))
}

func setupHandler(middleware mux.MiddlewareFunc, expectedValue string) http.Handler {
	router := mux.NewRouter()
	router.Use(middleware)
	router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		actualValue, err := namespace.FromContext(req.Context())

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if actualValue != expectedValue {
			http.Error(w, fmt.Sprintf("different ns in context: actual: %s; expected: %s", actualValue, expectedValue), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	return router
}
