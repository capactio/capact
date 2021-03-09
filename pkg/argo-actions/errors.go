package argoactions

import "github.com/pkg/errors"

func ErrMissingTypeInstanceValue(typeInstanceName string) error {
	return errors.Errorf("missing file with values for TypeInstances %s", typeInstanceName)
}

func ErrMissingResourceVersion() error {
	return errors.Errorf("resourceVersion is missing")
}
