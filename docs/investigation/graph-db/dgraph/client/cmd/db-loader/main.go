package main

import (
	"context"
	"log"

	"github.com/Project-Voltron/voltron/docs/investigation/graph-db/dgraph/client/internal"
	"github.com/Project-Voltron/voltron/docs/investigation/graph-db/dgraph/client/internal/client"

	"github.com/dgraph-io/dgo/v200/protos/api"
)

func main() {
	ochDir := "/Users/mszostok/workspace/go/src/github.com/Project-Voltron/voltron/docs/investigation/graph-db/dgraph/assets/och-content/"
	cli, err := client.New()
	check(err)

	//mu := api.Mutation{
	//	CommitNow: true,
	//	SetJson: []byte(`{
	//		"dgraph.type": "User",
	//		"uid": "_:us",
	//		"User.screen_name": "hakuna matata v222222222111111",
	//		"User.followers": 1993,
	//		"User.tweets": [{
	//		  "dgraph.type": "Tweets",
	//		  "Tweets.text": "sample heheheh v2222222221111",
	//		  "Tweets.timestamp": "1985-04-12T23:20:50.52Z",
	//		  "Tweets.users": [
	//			{
	//				"dgraph.type": "User",
	//				"uid": "_:us"
	//			}
	//		  ]
	//		}]
	//	}`),
	//}
	//resp, err := cli.NewTxn().Mutate(context.TODO(), &mu)
	//check(err)
	//fmt.Println(resp.Json)

	//return
	err = cli.Alter(context.TODO(), &api.Operation{DropAll: true})
	check(err)
	log.Println("Database dropped")

	internal.MustInitSchema("/Users/mszostok/workspace/go/src/github.com/Project-Voltron/voltron/docs/investigation/graph-db/dgraph/assets/public-och-schema.graphql")
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
