package clusterpolicy

import (
	"fmt"
	"strings"
)

type UnsupportedAPIVersionError struct {
	supported []string
}

func (e UnsupportedAPIVersionError) Error() string {
	return fmt.Sprintf(
		"unsupported API version; supported: [%s]",
		strings.Join(e.supported, ", "),
	)
}

func NewUnsupportedAPIVersionError(supported []string) *UnsupportedAPIVersionError {
	return &UnsupportedAPIVersionError{supported: supported}
}
