package public

import (
	"context"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/internal/regexutil"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"

	"github.com/pkg/errors"
)

// listAdditionalRefsFields defines preset for response fields required for collection Type's spec.additionalRefs
const listAdditionalRefsFields = TypeRevisionRootFields | TypeRevisionSpecAdditionalRefsField

// ListAdditionalRefsClient defines external Hub calls used by ListAdditionalRefs.
type ListAdditionalRefsClient interface {
	ListTypes(ctx context.Context, opts ...TypeOption) ([]*gqlpublicapi.Type, error)
}

// ListAdditionalRefsOutput holds Type's spec.additionalRefs entry indexed by TypeRef key.
type ListAdditionalRefsOutput map[types.TypeRef][]string

// ListAdditionalRefs knows how to fetch Type's spec.additionalRefs entries for all revisions for all given Types in a bit efficient way:
// - uses OR to get all Types in a single call
// - requests only required fields to prevent over-fetching
// - uses map type to give O(1) access to returned responses.
func ListAdditionalRefs(ctx context.Context, cli ListAdditionalRefsClient, reqTypes []types.TypeRef) (ListAdditionalRefsOutput, error) {
	filter := regexutil.OrStringSlice(mapToPaths(reqTypes))

	opts := []TypeOption{
		WithTypeRevisions(listAdditionalRefsFields),
		WithTypeFilter(gqlpublicapi.TypeFilter{
			PathPattern: ptr.String(filter),
		}),
	}

	res, err := cli.ListTypes(ctx, opts...)
	if err != nil {
		return ListAdditionalRefsOutput{}, errors.Wrap(err, "while fetching Types' additionalRefs for all revisions")
	}

	out := ListAdditionalRefsOutput{}
	for _, item := range res {
		if item == nil {
			continue
		}
		for _, rev := range item.Revisions {
			if rev.Spec == nil {
				continue
			}
			out[types.TypeRef{
				Path:     item.Path,
				Revision: rev.Revision,
			}] = rev.Spec.AdditionalRefs
		}
	}

	return out, nil
}

func mapToPaths(in []types.TypeRef) []string {
	var paths []string

	for _, expType := range in {
		paths = append(paths, expType.Path)
	}

	return paths
}
