package manifest

import (
	"context"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

// ValidateBackendStorageSchema is just a hack to export the internal method for testing purposes.
// The *_test.go files are not compiled into final binary, and as it's under _test.go it's also not accessible for other non-testing packages.
func (v *RemoteTypeValidator) ValidateBackendStorageSchema(ctx context.Context, entity types.Type) (ValidationResult, error) {
	return v.validateBackendStorageSchema(ctx, entity)
}
