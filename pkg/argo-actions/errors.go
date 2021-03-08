package argoactions

import "github.com/pkg/errors"

func ErrMissingTypeInstanceValue(typeInstanceName string) error {
	return errors.Errorf("missing value for TypeInstance %s", typeInstanceName)
}

func ErrMissingResourceVersion() error {
	return errors.Errorf("resourceVersion is missing")
}
