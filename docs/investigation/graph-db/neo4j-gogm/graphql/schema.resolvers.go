package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	go_cypherdsl "github.com/mindstand/go-cypherdsl"
	"log"
)

func (r *queryResolver) InterfaceGroups(ctx context.Context, filter *InterfaceGroupFilter) ([]*InterfaceGroup, error) {
	var items []*InterfaceGroup

	err := r.sess.LoadAllDepth(&items, 20)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return items, nil
}


//
// Not used resolvers - uncomment in `config.yaml` custom resolvers to play around
//

// Goal: MATCH (i:Interface {uuid: "be36d93f-b6ae-4b34-a321-4bdf4df9253f"})-[r:revision]->(ir:InterfaceRevision {revision: "0.0.1"}) RETURN ir
// How to achieve that? There is no documentation for go-cypherdsl
func (r *interfaceResolver) Revision(ctx context.Context, obj *Interface, revision string) (*InterfaceRevision, error) {
	condition := go_cypherdsl.C(&go_cypherdsl.ConditionConfig{
		Name:              "rev",
		Field:             "revision",
		ConditionOperator: go_cypherdsl.EqualToOperator,
		Check:             revision,
	}).And(&go_cypherdsl.ConditionConfig{}) // TODO: How to build the condition?

	var items []*InterfaceRevision
	err := r.sess.LoadAllDepthFilter(&items, 20, condition, map[string]interface{}{
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if len(items) > 1 {
		log.Println("more than one items")
		return nil, err
	}

	return items[0], nil
}

// Goal: MATCH (i:Interface {uuid: "be36d93f-b6ae-4b34-a321-4bdf4df9253f"})-[r:revision]->(ir:InterfaceRevision) RETURN ir
// How to achieve that? There is no documentation for go-cypherdsl
func (r *interfaceResolver) Revisions(ctx context.Context, obj *Interface) ([]*InterfaceRevision, error) {
	condition := go_cypherdsl.C(&go_cypherdsl.ConditionConfig{}) // TODO: How to build the condition?

	var items []*InterfaceRevision
	err := r.sess.LoadAllDepthFilter(&items, 20, condition, map[string]interface{}{})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if len(items) > 1 {
		log.Println("more than one items")
		return nil, err
	}

	return items, nil
}

// Interface returns InterfaceResolver implementation.
func (r *Resolver) Interface() InterfaceResolver { return &interfaceResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type interfaceResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
