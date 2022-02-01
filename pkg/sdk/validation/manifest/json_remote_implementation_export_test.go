package manifest

import (
	"context"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

// CheckParentNodesAssociation is just a hack to export the internal method for testing purposes.
// The *_test.go files are not compiled into final binary, and as it's under _test.go it's also not accessible for other non-testing packages.
func (v *RemoteImplementationValidator) CheckParentNodesAssociation(ctx context.Context, relations ParentNodesAssociation) (ValidationResult, error) {
	return v.checkParentNodesAssociation(ctx, relations)
}

//ValidateInputArtifactsNames exports validateInputArtifactsNames method for testing purposes.
func (v *RemoteImplementationValidator) ValidateInputArtifactsNames(ctx context.Context, entity types.Implementation) (ValidationResult, error) {
	return v.validateInputArtifactsNames(ctx, entity)
}
