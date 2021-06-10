package policy

import "github.com/pkg/errors"

var ErrCannotConvertAdditionalInput = errors.New("additional input cannot be converted to map[string]interface{}")
