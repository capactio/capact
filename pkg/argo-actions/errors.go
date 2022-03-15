package argoactions

import "github.com/pkg/errors"

// ErrMissingTypeInstanceValue returns an error indicating missing TypeInstance value file.
func ErrMissingTypeInstanceValue(typeInstanceName string) error {
	return errors.Errorf("missing file with values for TypeInstances %s", typeInstanceName)
}

// ErrMissingResourceVersion returns an error indicating missing resourceVersion.
func ErrMissingResourceVersion() error {
	return errors.Errorf("resourceVersion is missing")
}
