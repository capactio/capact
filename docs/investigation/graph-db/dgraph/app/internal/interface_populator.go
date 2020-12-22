package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/dgraph-io/dgo/v200"
	"github.com/dgraph-io/dgo/v200/protos/api"
	"github.com/mitchellh/mapstructure"
)

const interfaceDirName = "interface"

func MustPopulateInterfaces(cli *dgo.Dgraph, ochDir string) {
	interfacePath := filepath.Join(ochDir, interfaceDirName)
	iDirs, err := ioutil.ReadDir(interfacePath)
	requireNoErr(err)

	for _, iDir := range iDirs {
		if !iDir.IsDir() {
			continue
		}
		iGroupDir := filepath.Join(interfacePath, iDir.Name())
		iGroupFileName := fmt.Sprintf("%s.json", iDir.Name()) // - simplicity (named same as folder)

		iGroupID := loadInterfaceGroup(cli, iGroupDir, iGroupFileName)

		loadInterfaceRevisions(cli, iGroupID, iGroupDir, iGroupFileName)
	}
}

func loadInterfaceGroup(cli *dgo.Dgraph, iGroupDir, iGroupFileName string) string {
	iGroup, err := ioutil.ReadFile(filepath.Join(iGroupDir, iGroupFileName))
	requireNoErr(err)

	raw := map[string]interface{}{}
	err = json.Unmarshal(iGroup, &raw)
	requireNoErr(err)

	raw["uid"] = "_:iGroup"
	pb, err := json.Marshal(raw)
	requireNoErr(err)

	mu := &api.Mutation{
		CommitNow: true,
		SetJson:   pb,
	}

	response, err := cli.NewTxn().Mutate(context.TODO(), mu)
	requireNoErr(err)

	iGroupID := response.Uids["iGroup"]

	log.Printf("Inserted InterfaceGroup from file %v [id %s]", iGroupFileName, iGroupID)

	return iGroupID
}

func loadInterfaceRevisions(cli *dgo.Dgraph, iGroupID, iGroupDir, iGroupFileName string) {
	interfaces, err := ioutil.ReadDir(iGroupDir)
	requireNoErr(err)

	for _, i := range interfaces {
		if i.IsDir() || i.Name() == iGroupFileName {
			continue
		}

		interfaceToUpload, err := ioutil.ReadFile(filepath.Join(iGroupDir, i.Name()))
		requireNoErr(err)

		raw := map[string]interface{}{}
		err = json.Unmarshal(interfaceToUpload, &raw)
		requireNoErr(err)

		raw["uid"] = "_:inter"

		interfaceRev := InterfaceRevision{}
		err = mapstructure.Decode(raw, &interfaceRev)
		requireNoErr(err)

		inter := Interface{
			Uid:            "uid(existingInterface)", // if existingInterface is empty then default to blank node, so new id is created (https://dgraph.io/docs/mutations/blank-nodes/)
			Path:           interfaceRev.Metadata.Path,
			Name:           interfaceRev.Metadata.Name,
			Prefix:         interfaceRev.Metadata.Prefix,
			LatestRevision: raw,
			Revisions: []map[string]interface{}{
				raw,
			},
			DType: []string{"Interface"},
		}

		iGroup := InterfaceGroup{
			Uid:        iGroupID,
			DType:      []string{"InterfaceGroup"},
			Interfaces: []Interface{inter},
		}

		updateLatestRev, err := json.Marshal(iGroup)
		requireNoErr(err)

		iGroup.Interfaces[0].LatestRevision = nil
		doNotUpdateLatestRev, err := json.Marshal(iGroup)
		requireNoErr(err)

		q1 := `query GetInterfaces($path: string, $revision: string) {
				 latestRevIsLower as latestRev(func: type(Interface)) @filter(eq(Interface.path, $path)) @cascade {
					Interface.latestRevision @filter(lt(InterfaceRevision.revision, $revision)) {
					  InterfaceRevision.revision
					  uid
					}
				  }
			
				  existingInterface as inter(func: type(Interface)) @filter(eq(Interface.path, $path)) {
					uid
				  }
			  }`

		// TODO: think if we need to detect same InterfaceRevisions, sth like deepEqual and if same then do not insert?
		req := &api.Request{
			CommitNow: true,
			Query:     q1,
			Vars: map[string]string{
				"$path":     interfaceRev.Metadata.Path,
				"$revision": interfaceRev.Revision,
			},
			Mutations: []*api.Mutation{
				{
					// If found that the latest revision is lower
					Cond:    ` @if(eq(len(latestRevIsLower), 1) AND eq(len(existingInterface), 1))`,
					SetJson: updateLatestRev,
				},
				{
					// If found that Interface does not exist so we are inserting new object
					Cond:    ` @if(eq(len(existingInterface), 0))`,
					SetJson: updateLatestRev,
				},
				{
					Cond:    ` @if(eq(len(latestRevIsLower), 0) AND eq(len(existingInterface), 1)) `,
					SetJson: doNotUpdateLatestRev,
				},
			},
		}

		res, err := cli.NewTxn().Do(context.TODO(), req)
		requireNoErr(err)

		log.Printf("Inserted %v:%v, res: %v", interfaceRev.Metadata.Path, interfaceRev.Revision, string(res.Json))
	}
}
