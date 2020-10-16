package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/vrischmann/envconfig"
)

// Config holds application related configuration
type Config struct {
	Port int
}

func main() {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading configuration")

	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte("GraphQL Voltron Gateway - OK")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Starting server on %s", addr)

	err = http.ListenAndServe(addr, nil)
	exitOnError(err, "while starting HTTP server")
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
