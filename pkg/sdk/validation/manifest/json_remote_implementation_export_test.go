package manifest

import "context"

// CheckParentNodesAssociation is just a hack to export the internal method for testing purposes.
// The *_test.go files are not compiled into final binary, and as it's under _test.go it's also not accessible for other non-testing packages.
func (v *RemoteImplementationValidator) CheckParentNodesAssociation(ctx context.Context, relations ParentNodesAssociation) (ValidationResult, error) {
	return v.checkParentNodesAssociation(ctx, relations)
}
