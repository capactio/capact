package typeinstance

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	mockedresolver "projectvoltron.dev/voltron/internal/och/graphql/local/mocked-resolver"
	gqllocalapi "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
)

type TypeInstanceResolver struct {
	typeInstances []*gqllocalapi.TypeInstance
}

func NewResolver() *TypeInstanceResolver {
	return &TypeInstanceResolver{}
}

// init() is called in every method as there is no better place to call it
// NewResolver() does not return error so it cannot be used there
func (r *TypeInstanceResolver) init() error {
	if r.typeInstances != nil {
		return nil
	}
	typeInstances, err := mockedresolver.MockedTypeInstances()
	if err != nil {
		return err
	}
	r.typeInstances = append(r.typeInstances, typeInstances...)
	return nil
}

func (r *TypeInstanceResolver) TypeInstances(ctx context.Context, filter *gqllocalapi.TypeInstanceFilter) ([]*gqllocalapi.TypeInstance, error) {
	err := r.init()
	if err != nil {
		return []*gqllocalapi.TypeInstance{}, err
	}

	return r.typeInstances, nil
}

func (r *TypeInstanceResolver) TypeInstance(ctx context.Context, id string) (*gqllocalapi.TypeInstance, error) {
	err := r.init()
	if err != nil {
		return nil, err
	}
	for _, typeInstance := range r.typeInstances {
		if typeInstance.Metadata.ID == id {
			return typeInstance, nil
		}
	}
	return nil, nil
}

func (r *TypeInstanceResolver) CreateTypeInstance(ctx context.Context, in *gqllocalapi.CreateTypeInstanceInput) (*gqllocalapi.TypeInstance, error) {
	err := r.init()
	if err != nil {
		return nil, err
	}
	var revision string
	if in.TypeRef.Revision != nil {
		revision = *in.TypeRef.Revision
	} else {
		revision = ""
	}

	tags := []*gqllocalapi.TagReference{}
	for _, tag := range in.Tags {
		var revision string
		if tag.Revision != nil {
			revision = *tag.Revision
		} else {
			revision = ""
		}

		tags = append(tags, &gqllocalapi.TagReference{
			Path:     tag.Path,
			Revision: revision,
		})
	}

	newTypeInstance := &gqllocalapi.TypeInstance{
		Metadata: &gqllocalapi.TypeInstanceMetadata{
			ID:   uuid.New().String(),
			Tags: tags,
		},
		Spec: &gqllocalapi.TypeInstanceSpec{
			TypeRef: &gqllocalapi.TypeReference{
				Path:     in.TypeRef.Path,
				Revision: revision,
			},
			Value: in.Value,
		},
		ResourceVersion: 1,
	}

	r.typeInstances = append(r.typeInstances, newTypeInstance)
	return newTypeInstance, nil
}

func (r *TypeInstanceResolver) UpdateTypeInstance(ctx context.Context, id string, in *gqllocalapi.UpdateTypeInstanceInput) (*gqllocalapi.TypeInstance, error) {
	err := r.init()
	if err != nil {
		return nil, err
	}
	var typeInstance *gqllocalapi.TypeInstance
	for _, typeInstance = range r.typeInstances {
		if typeInstance.Metadata.ID == id {
			break
		}
	}
	if typeInstance == nil {
		return nil, fmt.Errorf("No TypeInstance with Id %s", id)
	}
	if typeInstance.ResourceVersion != in.ResourceVersion {
		return nil, fmt.Errorf("Wrong ResourceVersion for TypeInstance with Id %s, please use latest revision", id)
	}
	typeInstance.ResourceVersion++
	if in.TypeRef != nil {
		typeInstance.Spec.TypeRef.Path = in.TypeRef.Path
		if in.TypeRef.Revision != nil {
			typeInstance.Spec.TypeRef.Revision = *in.TypeRef.Revision
		}
	}
	typeInstance.Spec.Value = in.Value
	return typeInstance, nil
}

func (r *TypeInstanceResolver) DeleteTypeInstance(ctx context.Context, id string) (*gqllocalapi.TypeInstance, error) {
	err := r.init()
	if err != nil {
		return nil, err
	}

	index := -1
	var typeInstance *gqllocalapi.TypeInstance
	var i int
	for i, typeInstance = range r.typeInstances {
		if typeInstance.Metadata.ID == id {
			index = i
			break
		}
	}
	if index == -1 {
		return nil, nil
	}
	r.typeInstances[index] = r.typeInstances[len(r.typeInstances)-1]
	r.typeInstances[len(r.typeInstances)-1] = nil
	r.typeInstances = r.typeInstances[:len(r.typeInstances)-1]
	return typeInstance, nil
}
