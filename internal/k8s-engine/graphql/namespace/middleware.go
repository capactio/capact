package namespace

import "net/http"

const DefaultNamespace = "default"

type Middleware struct {
	headerName string
}

func NewMiddleware(headerName string) *Middleware {
	return &Middleware{headerName: headerName}
}

// Handle reads namespace from header and passes it to next handlers in request context.
func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ns := r.Header.Get(m.headerName)
			if ns == "" {
				ns = DefaultNamespace
			}

			ctx := SaveToContext(r.Context(), ns)

			next.ServeHTTP(w, r.WithContext(ctx))
		},
	)
}
