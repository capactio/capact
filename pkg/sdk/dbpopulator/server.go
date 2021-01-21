package dbpopulator

import (
	"io/ioutil"
	"log"
	"net/http"

	"sigs.k8s.io/yaml"
)

// TODO https://stackoverflow.com/questions/39320025/how-to-stop-http-listenandserve
func ServeJson(validPaths []string) {
	http.HandleFunc("/", jsonHandler(validPaths))
	srv := http.Server{Addr: "0.0.0.0:8080"}
	go srv.ListenAndServe()
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
