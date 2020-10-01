package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/vrischmann/envconfig"
)

// Config holds application related configuration
type Config struct {
	OCHMode string
	Port    int
}

var (
	ochMode = map[string]string{
		"local":  "OCH Local - OK",
		"public": "OCH Public - OK",
	}
	errWrongOchMode = fmt.Errorf("OCH mode needs to be specified. Possible options: %s", strings.Join(keys(ochMode), ", "))
)

func main() {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading configuration")

	msg, found := ochMode[cfg.OCHMode]
	if !found {
		exitOnError(errWrongOchMode, "while validating OCH mode")
	}

	http.HandleFunc("/statusz", func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte(msg)); err != nil {
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

func keys(in map[string]string) []string {
	var out []string
	for k := range in {
		out = append(out, k)
	}
	return out
}
