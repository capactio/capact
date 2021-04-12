package main

import (
	"context"
	"fmt"

	grakn "capact.io/capact/poc/graph-db/grakn/go-grakn/gograkn/session"
	"capact.io/capact/poc/graph-db/grakn/graphql"
)

type MyResolver struct {
}

func NewRootResolver() *MyResolver {
	return &MyResolver{}
}

func (r *MyResolver) Query() graphql.QueryResolver {
	return r
}

func (r *MyResolver) InterfaceGroups(ctx context.Context, filter *graphql.InterfaceGroupFilter) ([]*graphql.InterfaceGroup, error) {
	preloads := GetPreloads(ctx)
	q := toQuery(preloads)

	res, err := query(q)
	if err != nil {
		return []*graphql.InterfaceGroup{}, err
	}

	fmt.Printf("%+v", res)
	return []*graphql.InterfaceGroup{toInterfaceGroup(res, preloads)}, nil
}

func toQuery(fields []string) string {
	query := "match $ifaceGroup isa interfaceGroup"
	relations := ""
	mapped := map[string]bool{}
	for _, k := range fields {
		mapped[k] = true
	}

	if _, ok := mapped["interfaces"]; ok {
		query += ";$iface isa interface"
		relations += ";$gr1 (groups: $ifaceGroup, grouped: $iface) isa grouping"

		if _, ok := mapped["interfaces.name"]; ok {
			query += ", has name $ifaceName"
		}

		// this would go to a new funciton which would get implementation query
		if _, ok := mapped["interfaces.revisions.implementations"]; ok {
			query += ";$impl isa implementation"
			relations += ";$impl-iface (defines: $iface, implements: $impl) isa implementator"
		}

		if _, ok := mapped["interfaces.revisions.implementations.name"]; ok {
			query += ", has name $implName"
		}
	}
	query += relations + ";get;"

	fmt.Printf("\n%s\n\n", query)
	return query
}

///XXX ignores relations
func toInterfaceGroup(concepts []map[string]*grakn.Concept, fields []string) *graphql.InterfaceGroup {
	ig := &graphql.InterfaceGroup{}

	mapped := map[string]bool{}
	for _, k := range fields {
		mapped[k] = true
	}

	if _, ok := mapped["interfaces"]; ok {
		interfaces := []*graphql.Interface{}
		for _, m := range concepts {
			if v, ok := m["ifaceName"]; ok {
				iface := &graphql.Interface{}
				iface.Name = v.ValueRes.Value.GetString_()
				interfaces = append(interfaces, iface)
			}
		}
		ig.Interfaces = append(ig.Interfaces, interfaces...)

	}

	return ig
}
