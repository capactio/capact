package repometadata

import (
	"context"

	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type RepoMetadataResolver struct{}

func NewResolver() *RepoMetadataResolver {
	return &RepoMetadataResolver{}
}

func (r *RepoMetadataResolver) RepoMetadata(ctx context.Context) (*gqlpublicapi.RepoMetadata, error) {
	return dummyRepoMetadata(), nil
}

func (r *RepoMetadataResolver) Revision(ctx context.Context, obj *gqlpublicapi.RepoMetadata, revision string) (*gqlpublicapi.RepoMetadataRevision, error) {
	return &gqlpublicapi.RepoMetadataRevision{}, nil
}

func dummyRepoMetadata() *gqlpublicapi.RepoMetadata {
	return &gqlpublicapi.RepoMetadata{
		Name:   "metadata",
		Prefix: "cap.core",
		Path:   "cap.core.metadata",
	}
}
