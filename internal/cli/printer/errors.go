package printer

import "fmt"

func PrintErrors(errs []error) {
	for _, err := range errs {
		fmt.Printf("%s\n", err.Error())
	}
}
