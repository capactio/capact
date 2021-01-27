package dbpopulator

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"

	"sigs.k8s.io/yaml"
)

// ServeJson serves OCH Manifests
// manifests are converted from YAML to JSON when requested
func ServeJson(ctx context.Context, validPaths []string) {
	http.HandleFunc("/", jsonHandler(validPaths))
	srv := http.Server{Addr: "0.0.0.0:8080"}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	select {
	case <-ctx.Done():
		srv.Shutdown(ctx)
	}
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
			log.Fatal(err)
		}
		converted, err := yaml.YAMLToJSON(content)
		w.Write(converted)
	}
}
