package policy

import "github.com/pkg/errors"

// ErrPolicyConfigMapNotFound returns an error indicating that Policy cannot be found.
var ErrPolicyConfigMapNotFound = errors.New("ConfigMap with Policy not found")
