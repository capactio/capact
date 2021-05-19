package printer

import (
	"fmt"
	"os"
)

func PrintErrors(errs []error) {
	for _, err := range errs {
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		}
	}
}
