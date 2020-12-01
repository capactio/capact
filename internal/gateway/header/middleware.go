package header

import (
	"net/http"

	"github.com/nautilus/graphql"
)

func SaveHeadersInCtxHTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := NewContext(r.Context(), r.Header)
			next.ServeHTTP(w, r.WithContext(ctx))
		},
	)
}

func SetHeadersFromCtxGQLMiddleware() graphql.NetworkMiddleware {
	return func(request *http.Request) error {
		headers, ok := FromContext(request.Context())
		if !ok {
			return nil
		}

		for key, values := range headers {
			if _, ok := request.Header[key]; ok {
				// Do not override any existing headers
				continue
			}

			request.Header[key] = values
		}

		return nil
	}
}
