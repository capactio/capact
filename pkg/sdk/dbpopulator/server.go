package dbpopulator

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

// ServeJson serves OCH Manifests
// manifests are converted from YAML to JSON when requested
func MustServeJSON(ctx context.Context, listenPort int, validPaths []string) {
	http.HandleFunc("/", jsonHandler(validPaths))
	srv := http.Server{Addr: fmt.Sprintf("0.0.0.0:%d", listenPort)}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	<-ctx.Done()
	_ = srv.Shutdown(ctx)
}

func jsonHandler(validPaths []string) func(http.ResponseWriter, *http.Request) {
	// for faster lookup...
	paths := map[string]struct{}{}
	for _, path := range validPaths {
		paths[path] = struct{}{}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		_, ok := paths[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		content, err := ioutil.ReadFile(r.URL.Path)
		if err != nil {
			errMsg := errors.Wrapf(err, "cannot read %s", r.URL.Path).Error()
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		converted, err := yaml.YAMLToJSON(content)
		if err != nil {
			errMsg := errors.Wrapf(err, "cannot convert %s to JSON", r.URL.Path).Error()
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		_, err = w.Write(converted)
		if err != nil {
			errMsg := errors.Wrapf(err, "cannot write response for %s", r.URL.Path).Error()
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
	}
}
