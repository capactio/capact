package main

import (
	"context"
	"log"

	"capact.io/capact/docs/investigation/graph-db/dgraph/app/internal"
	"capact.io/capact/docs/investigation/graph-db/dgraph/app/internal/client"

	"github.com/dgraph-io/dgo/v200/protos/api"
)

func main() {
	ochDir := "../assets/och-content/"
	cli, err := client.New()
	check(err)

	err = cli.Alter(context.TODO(), &api.Operation{DropAll: true})
	check(err)
	log.Println("Database dropped")

	internal.MustInitSchema("../assets/public-och-schema.graphql")
	log.Println("New schema loaded")

	internal.MustPopulateType(cli, ochDir)

	internal.MustPopulateInterfaces(cli, ochDir)
	log.Println("Interfaces populated")

	internal.MustPopulateImplementations(cli, ochDir)
	log.Println("Implementations populated")
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
