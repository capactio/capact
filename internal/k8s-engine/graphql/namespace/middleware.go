package namespace

import "net/http"

const (
	// NamespaceHeaderName defines HTTP header name where Kubernetes Namespace is stored.
	NamespaceHeaderName = "NAMESPACE"
	// DefaultNamespace defines default Kubernetes Namespace name.
	DefaultNamespace = "default"
)

// Middleware provides functionality to handle Namespace property in HTTP requests.
type Middleware struct {
	headerName string
}

// NewMiddleware returns a new Middleware instance.
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
