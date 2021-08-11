package action

import (
	gqlengine "capact.io/capact/pkg/engine/api/graphql"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

func convertTypeInstancesRefsToGQL(refs []types.InputTypeInstanceRef) []*gqlengine.InputTypeInstanceData {
	var out []*gqlengine.InputTypeInstanceData
	for idx := range refs {
		gql := gqlengine.InputTypeInstanceData(refs[idx])
		out = append(out, &gql)
	}
	return out
}

func convertParametersToGQL(parameters string) *gqlengine.JSON {
	if parameters == "" {
		return nil
	}
	out := gqlengine.JSON(parameters)
	return &out
}
