package scalar

import "github.com/pkg/errors"

func ConvertToString(in interface{}) (string, error) {
	if in == nil {
		return "", errors.New("input should not be nil")
	}

	value, ok := in.(string)
	if !ok {
		return "", errors.Errorf("unexpected input type: %T, should be string", in)
	}

	return value, nil
}
