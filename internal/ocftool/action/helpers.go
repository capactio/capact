package action

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
)

func isDNSSubdomain(val interface{}) error {
	if str, ok := val.(string); ok {
		validation.IsDNS1123Subdomain(str)
		if msgs := validation.IsDNS1123Subdomain(str); len(msgs) != 0 {
			return fmt.Errorf("%s", strings.Join(msgs, ", "))
		}
	} else {
		// otherwise we cannot convert the value into a string and cannot enforce length
		return fmt.Errorf("cannot enforce DNS syntax validation on response of type %v", reflect.TypeOf(val).Name())
	}

	// the input is fine
	return nil
}
