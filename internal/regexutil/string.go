package regexutil

import (
	"fmt"
	"strings"
)

// OrStringSlice returns or regexp for all items in a given slice.
func OrStringSlice(in []string) string {
	return fmt.Sprintf(`(%s)`, strings.Join(in, "|"))
}
