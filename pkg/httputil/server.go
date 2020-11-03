package httputil

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// StartableServer represents server which can be started on demand.
type StartableServer interface {
	// Start the server and blocks until the channel is closed or an error occurs.
	// MUST shutdown gracefully the server when channel is closed.
	Start(stop <-chan struct{}) error
}

var _ StartableServer = &Server{}

// Server provides functionality to create and start the HTTP server.
type Server struct {
	mux  http.Handler
	addr string
	log  *zap.Logger
}

// NewStartableServer returns new Server instance.
func NewStartableServer(log *zap.Logger, addr string, handler http.Handler) *Server {
	return &Server{
		mux:  handler,
		addr: addr,
		log:  log,
	}
}

// Start the HTTP server and blocks until the channel is closed or an error occurs.
func (svr *Server) Start(stop <-chan struct{}) error {
	srv := &http.Server{Addr: svr.addr, Handler: svr.mux}
	go func() {
		<-stop
		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			svr.log.Error("shutting down HTTP server", zap.Error(err))
		}
	}()

	svr.log.Info("Starting HTTP server", zap.String("addr", svr.addr))
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return errors.Wrap(err, "while starting HTTP server")
	}

	return nil
}
