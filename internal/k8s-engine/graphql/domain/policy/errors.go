package policy

import "github.com/pkg/errors"

// ErrCannotConvertAdditionalInput returns an error indicating that converting additional input was not possible.
var ErrCannotConvertAdditionalInput = errors.New("additional input cannot be converted to map[string]interface{}")
