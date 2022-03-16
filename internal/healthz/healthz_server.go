package healthz

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"capact.io/capact/pkg/httputil"
)

// NewHTTPServer returns new HTTP server with preconfigured `/healthz` endpoint.
func NewHTTPServer(namedLogger *zap.Logger, healthzAddr, appName string) httputil.StartableServer {
	router := mux.NewRouter()
	router.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		if _, err := fmt.Fprintf(w, "%s - OK", appName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	return httputil.NewStartableServer(
		namedLogger.With(zap.String("server", "healthz")),
		healthzAddr,
		router,
	)
}
