package namespace

import "net/http"

const (
	NamespaceHeaderName = "NAMESPACE"
	DefaultNamespace    = "default"
)

type Middleware struct {
	headerName string
}

func NewMiddleware() *Middleware {
	return &Middleware{headerName: NamespaceHeaderName}
}

// Handle reads namespace from header and passes it to next handlers in request context.
func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ns := r.Header.Get(m.headerName)
			if ns == "" {
				ns = DefaultNamespace
			}

			ctx := NewContext(r.Context(), ns)

			next.ServeHTTP(w, r.WithContext(ctx))
		},
	)
}
