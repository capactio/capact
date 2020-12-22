package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/dgo/v200"
	"github.com/dgraph-io/dgo/v200/protos/api"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

const implementationDirName = "implementation"

func MustPopulateImplementations(cli *dgo.Dgraph, ochDir string) {
	implPath := filepath.Join(ochDir, implementationDirName)
	dirs, err := ioutil.ReadDir(implPath)
	requireNoErr(err)

	for _, implDir := range dirs {
		if !implDir.IsDir() {
			continue
		}

		specificImplDir := filepath.Join(implPath, implDir.Name())
		implementations, err := ioutil.ReadDir(specificImplDir)
		requireNoErr(err)

		for _, i := range implementations {
			if i.IsDir() {
				continue
			}

			implToUpload, err := ioutil.ReadFile(filepath.Join(specificImplDir, i.Name()))
			requireNoErr(err)

			raw := map[string]interface{}{}
			err = json.Unmarshal(implToUpload, &raw)
			requireNoErr(err)

			raw["uid"] = "_:impl"

			implRev := ImplementationRevision{}
			err = mapstructure.Decode(raw, &implRev)
			requireNoErr(err)

			interfacesUID := getImplementedInterfaceIds(cli, implRev)
			raw["ImplementationRevision.interfaces"] = interfacesUID

			impl := Implementation{
				Uid:            "uid(existingImpl)", // if existingImpl is empty then default to blank node, so new id is created (https://dgraph.io/docs/mutations/blank-nodes/)
				Path:           implRev.Metadata.Path,
				Name:           implRev.Metadata.Name,
				Prefix:         implRev.Metadata.Prefix,
				LatestRevision: raw,
				Revisions: []map[string]interface{}{
					raw,
				},
				DType: []string{"Implementation"},
			}

			updateLatestRev, err := json.Marshal(impl)
			requireNoErr(err)

			impl.LatestRevision = nil
			doNotUpdateLatestRev, err := json.Marshal(impl)
			requireNoErr(err)

			q1 := `query GetImplementations($path: string, $revision: string) {
				 latestRevIsLower as latestRev(func: type(Implementation)) @filter(eq(Implementation.path, $path)) @cascade {
					Implementation.latestRevision @filter(lt(ImplementationRevision.revision, $revision)) {
					  ImplementationRevision.revision
					  uid
					}
				  }
			
				  existingImpl as inter(func: type(Implementation)) @filter(eq(Implementation.path, $path)) {
					uid
				  }
			  }`

			// TODO: think if we need to detect same ImplementationRevisions, sth like deepEqual and if same then do not insert?
			req := &api.Request{
				CommitNow: true,
				Query:     q1,
				Vars: map[string]string{
					"$path":     implRev.Metadata.Path,
					"$revision": implRev.Revision,
				},
				Mutations: []*api.Mutation{
					{
						// If found that the latest revision is lower
						Cond:    ` @if(eq(len(latestRevIsLower), 1) AND eq(len(existingImpl), 1))`,
						SetJson: updateLatestRev,
					},
					{
						// If found that Impl does not exist so we are inserting new object
						Cond:    ` @if(eq(len(existingImpl), 0))`,
						SetJson: updateLatestRev,
					},
					{
						Cond:    ` @if(eq(len(latestRevIsLower), 0) AND eq(len(existingImpl), 1)) `,
						SetJson: doNotUpdateLatestRev,
					},
				},
			}

			res, err := cli.NewTxn().Do(context.TODO(), req)
			requireNoErr(err)

			log.Printf("Inserted %v:%v, res: %v", implRev.Metadata.Path, implRev.Revision, string(res.Json))
		}

	}
}

type UID struct {
	Uid   string   `json:"uid"`
	DType []string `json:"dgraph.type,omitempty"`
}
type ReferenceInterfaceRevision struct {
	UID
	ImplRevision UID `json:"InterfaceRevision.implementedBy"`
}

func getInterfacePaths(revs []InterfaceReference) string {
	var paths []string
	for idx := range revs {
		paths = append(paths, revs[idx].Path)
	}
	return strings.Join(paths, ",")
}

func getImplementedInterfaceIds(cli *dgo.Dgraph, rev ImplementationRevision) []ReferenceInterfaceRevision {
	// There is no for each yet: https://discuss.dgraph.io/t/foreach-func-in-dql-loops-in-bulk-upsert/5533/9
	// We could repeat the query for each path+revision pair, but we can also query all matching paths with all revisions and filter
	// them programmatically.
	//
	// Additionally, there is no support for slice input: https://discuss.dgraph.io/t/support-lists-in-query-variables-dgraphs-graphql-variable/8758
	qInterfaces := fmt.Sprintf(`
					{
					  interfaces(func: type(Interface)) @cascade @filter(eq(Interface.path, [%s])) {
						path: Interface.path
						revisions: Interface.revisions {
						  uid
						  rev: InterfaceRevision.revision
						}
					  }
					}
			`, getInterfacePaths(rev.Spec.Implements))
	resp, err := cli.NewTxn().Query(context.TODO(), qInterfaces)
	requireNoErr(err)

	var decode struct {
		Interfaces []DecodeInterfaceQuery
	}

	err = json.Unmarshal(resp.GetJson(), &decode)
	requireNoErr(err)

	indexed := map[string]DecodeInterfaceQuery{}

	for idx := range decode.Interfaces {
		indexed[decode.Interfaces[idx].Path] = decode.Interfaces[idx]
	}

	var ids []ReferenceInterfaceRevision

	// errors could be aggregated to get rid of fast return
	for _, implements := range rev.Spec.Implements {
		i, found := indexed[implements.Path]

		if !found {
			requireNoErr(fmt.Errorf("Interface was not found for %s", implements.Path))
		}

		var interfaceID string
		for _, rev := range i.Revisions {
			if rev.Rev == implements.Revision {
				interfaceID = rev.Uid
			}
		}

		if interfaceID == "" {
			requireNoErr(fmt.Errorf("Interface %q was not found for revision %s", implements.Path, implements.Revision))
		}
		ids = append(ids, ReferenceInterfaceRevision{
			UID: UID{
				Uid:   interfaceID,
				DType: []string{"InterfaceRevision"},
			},
			ImplRevision: UID{ // TODO: why we need to add also the id here? In GraphQL it is not needed
				Uid:   "_:impl",
				DType: []string{"ImplementationRevision"},
			},
		})
	}

	return ids
}
