// @generated - This was created as a part of investigation. We mark it as generate to exlude it from goreportcard to do not have missleading issues.:golint
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"capact.io/capact/docs/investigation/graph-db/dgraph/app/internal/client"

	"github.com/dgraph-io/dgo/v200"
	"github.com/gorilla/mux"
)

var findJsonKeys = regexp.MustCompile(`"[a-zA-Z.]*":`)

func main() {
	cli, err := client.New()
	check(err)

	handler := &ImplHandler{cli: cli}

	r := mux.NewRouter()
	r.Handle("/implementations", handler).Methods(http.MethodPost)

	log.Println("Start listening on :8888")
	check(http.ListenAndServe(":8888", r))
}

type ImplHandler struct {
	cli *dgo.Dgraph
}

type ImplementationFilter struct {
	InterfaceID string `json:"id"`
}

type DecodeImplementationsQuery struct {
	Uid  string
	Path string
	Rev  string
}

func (h *ImplHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var filter ImplementationFilter
	err := json.NewDecoder(r.Body).Decode(&filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	q := `query GetSatisfiedImpl($interfaceID: string) {
    implementations(func: type(ImplementationRevision)) @cascade @normalize {
      uid
      spec: ImplementationRevision.spec {
        ImplementationSpec.requires {
          ImplementationRequirement.oneOf {
            ImplementationRequirementItem.typeRef {
              path: TypeReference.path
              rev: TypeReference.revision
            }
          }
        }
      }
      ImplementationRevision.interfaces @filter(uid($interfaceID))
    }
}`
	resp, err := h.cli.NewTxn().QueryWithVars(r.Context(), q, map[string]string{
		"$interfaceID": filter.InterfaceID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var decode struct {
		Implementations []DecodeImplementationsQuery
	}

	err = json.Unmarshal(resp.GetJson(), &decode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ids := getOnlySatisfiedImplIDs(decode.Implementations)

	if len(ids) == 0 {
		w.Write([]byte("[]"))
		return
	}

	// The one option is to use aliases but it is problematic as we need to specify all possible key-value pairs
	//
	//qAllNodesWithCustomAliases := fmt.Sprintf(`{
	//	 AllNodes(func: type(Implementation)) @filter(uid_in(Implementation.latestRevision, [%s])) {
	//		name: Implementation.name
	//		path: Implementation.path
	//		prefix: Implementation.prefix
	//		latestRevision: Implementation.latestRevision {
	//		  metadata: ImplementationRevision.metadata {
	//			expand(_all_)
	//		  }
	//		  revision: ImplementationRevision.revision
	//
	//		}
	//
	//	  }
	//	}`, strings.Join(ids, ","))

	// There is no support for slice input: https://discuss.dgraph.io/t/support-lists-in-query-variables-dgraphs-graphql-variable/8758
	// SIMPLIFICATION: only the latest revision are filtered. The revisions entry is not taken into account.
	qAllNodes := fmt.Sprintf(`{
		 AllNodes(func: type(Implementation)) @filter(uid_in(Implementation.latestRevision, [%s])) {
			expand(_all_) {
			  expand(_all_) {
				expand(_all_) {
				  expand(_all_)
				}
			  }
			}
		  }
		}`, strings.Join(ids, ","))

	resp, err = h.cli.NewTxn().Query(r.Context(), qAllNodes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var decodeAllNodes struct {
		AllNodes json.RawMessage
	}
	err = json.Unmarshal(resp.GetJson(), &decodeAllNodes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	transformed := removeTypePrefixesFromJSONKeys(string(decodeAllNodes.AllNodes))
	w.Write([]byte(transformed))
}

// Benchmark: (we need to check how dgraph does it internally). Using regex is probably not a good idea at all.
//
// go test  -bench=. -benchmem -cpu 1 -benchtime 5s ./cmd/resolver-svr/...
//   goos: darwin
//   goarch: amd64
//   pkg: capact.io/capact/docs/investigation/graph-db/dgraph/app/cmd/resolver-svr
//   BenchmarkRemoveTypePrefixesFromJSONKeys            15492            403021 ns/op          139093 B/op        200 allocs/op
//   PASS
//   ok      capact.io/capact/docs/investigation/graph-db/dgraph/app/cmd/resolver-svr      10.420s
func removeTypePrefixesFromJSONKeys(in string) string {
	out := findJsonKeys.ReplaceAllStringFunc(in, func(match string) string {
		idx := strings.LastIndex(match, ".")
		if idx == -1 {
			return match
		}

		return `"` + match[idx+1:]
	})

	return out
}

// dummy - return all but we have access here for required types by implementation so we can execute some business logic here.
func getOnlySatisfiedImplIDs(in []DecodeImplementationsQuery) []string {
	var ids []string
	for idx := range in {
		ids = append(ids, in[idx].Uid)
	}

	return ids
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
