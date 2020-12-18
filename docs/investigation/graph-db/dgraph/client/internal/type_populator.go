package internal

import (
	"context"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/dgraph-io/dgo/v200"
	"github.com/dgraph-io/dgo/v200/protos/api"
)

const typeDirName = "type"

func MustPopulateType(cli *dgo.Dgraph, ochDir string) {
	typesPath := filepath.Join(ochDir, typeDirName)
	dirs, err := ioutil.ReadDir(typesPath)
	requireNoErr(err)

	for _, typeDir := range dirs {
		if !typeDir.IsDir() {
			continue
		}

		specificTypDir := filepath.Join(typesPath, typeDir.Name())
		types, err := ioutil.ReadDir(specificTypDir)
		requireNoErr(err)

		for _, i := range types {
			if i.IsDir() {
				continue
			}

			typeToUpload, err := ioutil.ReadFile(filepath.Join(specificTypDir, i.Name()))
			requireNoErr(err)

			res, err := cli.NewTxn().Mutate(context.TODO(), &api.Mutation{
				CommitNow: true,
				SetJson:   typeToUpload,
			})
			requireNoErr(err)

			log.Printf("Type Inserted, res: %v", string(res.Json))
		}

	}
}
