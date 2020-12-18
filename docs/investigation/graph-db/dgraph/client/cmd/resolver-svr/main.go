package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Project-Voltron/voltron/docs/investigation/graph-db/dgraph/client/internal/client"

	"github.com/dgraph-io/dgo/v200"
	"github.com/gorilla/mux"
)

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

	//ids := getOnlySatisfiedImplIDs(decode.Implementations)

	qAllNodes := `{
		 AllNodes(func: type(Implementation)) @filter(uid_in(Implementation.latestRevision, 0x5c7)) {
			dgraph.type
			expand(_all_) {
			  expand(_all_) {
				expand(_all_) {
				  expand(_all_)
				}
			  }
			}
			
		  }
		}`

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
	w.Write(decodeAllNodes.AllNodes)
}

// dummy - return all
func getOnlySatisfiedImplIDs(in []DecodeImplementationsQuery) []string {
	var ids []string
	for _, impl := range in {
		ids = append(ids, impl.Uid)
	}

	return ids
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
