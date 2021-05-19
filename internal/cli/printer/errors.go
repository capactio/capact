package printer

import (
	"fmt"
	"os"
)

func PrintErrors(errs []error) {
	for _, err := range errs {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
}
