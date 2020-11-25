package repometadata

import (
	"context"

	mockedresolver "projectvoltron.dev/voltron/internal/och/graphql/public/mocked-resolver"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type RepoMetadataResolver struct{}

func NewResolver() *RepoMetadataResolver {
	return &RepoMetadataResolver{}
}

func (r *RepoMetadataResolver) RepoMetadata(ctx context.Context) (*gqlpublicapi.RepoMetadata, error) {
	repo, err := mockedresolver.MockedRepoMetadata()
	if err != nil {
		return &gqlpublicapi.RepoMetadata{}, err
	}
	return repo, nil
}

func (r *RepoMetadataResolver) Revision(ctx context.Context, obj *gqlpublicapi.RepoMetadata, revision string) (*gqlpublicapi.RepoMetadataRevision, error) {
	for _, rev := range obj.Revisions {
		if rev.Revision == revision {
			return rev, nil
		}
	}
	return nil, nil
}
