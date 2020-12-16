package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/dgo/v200"
	"github.com/dgraph-io/dgo/v200/protos/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
	"log"
)

type Interface struct {
	Uid            string              `json:"uid,omitempty"`
	Path           string              `json:"Interface.path"`
	LatestRevision *InterfaceRevision  `json:"Interface.latestRevision,omitempty"`
	Revisions      []InterfaceRevision `json:"Interface.revisions"`
	DType          []string            `json:"dgraph.type,omitempty"`
}

type InterfaceRevision struct {
	Uid      string   `json:"uid,omitempty"`
	Revision string   `json:"InterfaceRevision.revision"`
	DType    []string `json:"dgraph.type,omitempty"`
}

func main() {
	cli := newClient()

	ctx := context.Background()

	txn := cli.NewTxn()
	defer txn.Discard(ctx)

	newLast := Interface{
		Uid:  "0x5",
		Path: "cap.interface.productivity.jira.install",
		LatestRevision: &InterfaceRevision{
			Uid:      "_:rev",
			Revision: "0.1.3",
			DType:    []string{"InterfaceRevision"},
		},
		Revisions: []InterfaceRevision{
			{
				Uid:      "_:rev",
				Revision: "0.1.3",
				DType:    []string{"InterfaceRevision"},
			},
		},
		DType: []string{"Interface"},
	}

	updateLatestRev, err := json.Marshal(newLast)
	check(err)

	newLast.LatestRevision = nil
	doNotUpdateLatestRev, err := json.Marshal(newLast)
	check(err)

	q1 := `
{
  u1 as inter(func: uid("0x5")) @filter(type(Interface)) @cascade {
    Interface.latestRevision @filter( lt(InterfaceRevision.revision, "0.1.3")){
     uid
    }
  }
}`

	req := &api.Request{
		CommitNow: true,
		Query:     q1,
		Mutations: []*api.Mutation{
			//{
			//	Cond:    ` @if(eq(late, "0.3.0"))`,
			//	SetJson: updateLatestRev,
			//},
			{
				Cond:    ` @if(eq(len(u1), 1777))`,
				SetJson: updateLatestRev,
			},
			{
				Cond:    ` @if(eq(len(u1), 0)) `,
				SetJson: doNotUpdateLatestRev,
			},
		},
	}
	res, err := txn.Do(ctx, req)
	check(err)

	//txn.Query()
	//err = txn.Commit(ctx)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//if err == dgo.ErrAborted {
	//	// Retry or handle error
	//}

	fmt.Println(string(res.Json))
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func newClient() *dgo.Dgraph {
	// Dial a gRPC connection. The address to dial to can be configured when
	// setting up the dgraph cluster.
	dialOpts := append([]grpc.DialOption{},
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)))
	d, err := grpc.Dial("localhost:9080", dialOpts...)

	if err != nil {
		log.Fatal(err)
	}

	return dgo.NewDgraphClient(
		api.NewDgraphClient(d),
	)
}
