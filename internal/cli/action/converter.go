package action

import (
	"encoding/json"

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

func convertParametersToGQL(parameters json.RawMessage) *gqlengine.JSON {
	if parameters == nil {
		return nil
	}

	out := gqlengine.JSON(string(parameters))
	return &out
}
