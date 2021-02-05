package argo

import (
	"context"

	ochpublicgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
)

func interfaceRefToOCH(in types.InterfaceRef) ochpublicgraphql.InterfaceReference {
	return ochpublicgraphql.InterfaceReference{
		Path:     in.Path,
		Revision: stringOrEmpty(in.Revision),
	}
}

func stringOrEmpty(in *string) string {
	if in != nil {
		return *in
	}
	return ""
}

func shouldExit(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
