// @generated - This was created as a part of investigation. We mark it as generate to exlude it from goreportcard to do not have missleading issues.:golint
package main

import (
	"context"
	"time"

	grakn "capact.io/capact/poc/graph-db/grakn/go-grakn/gograkn/session"
	"github.com/99designs/gqlgen/graphql"
	grpc "google.golang.org/grpc"
)

func GetPreloads(ctx context.Context) []string {
	return GetNestedPreloads(
		graphql.GetOperationContext(ctx),
		graphql.CollectFieldsCtx(ctx, nil),
		"",
	)
}

func GetNestedPreloads(ctx *graphql.OperationContext, fields []graphql.CollectedField, prefix string) (preloads []string) {
	for _, column := range fields {
		prefixColumn := GetPreloadString(prefix, column.Name)
		preloads = append(preloads, prefixColumn)
		preloads = append(preloads, GetNestedPreloads(ctx, graphql.CollectFields(ctx, column.Selections, nil), prefixColumn)...)
	}
	return
}

func GetPreloadString(prefix, name string) string {
	if len(prefix) > 0 {
		return prefix + "." + name
	}
	return name
}

func query(q string) ([]map[string]*grakn.Concept, error) {
	concepts := []map[string]*grakn.Concept{}

	address := ":48555"
	username := ""
	password := ""

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return concepts, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client := grakn.NewSessionServiceClient(conn)
	sessionRes, err := client.Open(ctx, &grakn.Session_Open_Req{Username: username, Password: password, Keyspace: "och"})
	if err != nil {
		return concepts, err
	}
	tsClient, err := client.Transaction(ctx)
	if err != nil {
		return concepts, err
	}
	openRq := &grakn.Transaction_Req_OpenReq{
		OpenReq: &grakn.Transaction_Open_Req{
			SessionId: sessionRes.SessionId,
			Type:      grakn.Transaction_READ,
		},
	}
	err = tsClient.Send(&grakn.Transaction_Req{Req: openRq})
	if err != nil {
		return concepts, err
	}

	tsRes, err := tsClient.Recv()
	if err != nil {
		return concepts, err
	}

	openRes := tsRes.GetOpenRes()
	if openRes == nil {
		return concepts, err
	}

	queryOptions := &grakn.Transaction_Query_Options{}
	query := &grakn.Transaction_Query_Iter_Req{Query: q, Options: queryOptions}
	request := &grakn.Transaction_Req{
		Req: &grakn.Transaction_Req_IterReq{
			IterReq: &grakn.Transaction_Iter_Req{
				Options: &grakn.Transaction_Iter_Req_Options{BatchSize: &grakn.Transaction_Iter_Req_Options_All{All: true}},
				Req: &grakn.Transaction_Iter_Req_QueryIterReq{
					QueryIterReq: query,
				},
			},
		},
	}

	err = tsClient.Send(request)
	if err != nil {
		return concepts, err
	}

	for {
		tsRes, err = tsClient.Recv()
		if err != nil {
			return concepts, err
		}
		iterRes := tsRes.GetIterRes()
		if iterRes == nil {
			return concepts, err
		}
		isDone := iterRes.GetDone()
		if isDone {
			break
		}
		concepts = append(concepts, iterRes.GetQueryIterRes().GetAnswer().GetConceptMap().GetMap())
	}
	return concepts, nil
}
