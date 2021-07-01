package printer

import (
	"fmt"
	"os"
)

// PrintErrors prints a given slice of error into standard error output.
// If slice is nil, doesn't print nothing.
func PrintErrors(errs []error) {
	for _, err := range errs {
		if err == nil {
			continue
		}
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
}
