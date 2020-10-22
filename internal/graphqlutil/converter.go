package graphqlutil

import "github.com/pkg/errors"

// ScalarToString tries to convert a given empty interface to string type.
// Returns error if conversion is not possible.
func ScalarToString(in interface{}) (string, error) {
	if in == nil {
		return "", errors.New("input should not be nil")
	}

	value, ok := in.(string)
	if !ok {
		return "", errors.Errorf("unexpected input type: %T, should be string", in)
	}

	return value, nil
}
