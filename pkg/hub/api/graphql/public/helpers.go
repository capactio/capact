package graphql

import (
	"fmt"
	"strings"

	"capact.io/capact/internal/ptr"
	"github.com/pkg/errors"
)

func NewActionImportsError(actionRef string) error {
	return errors.Errorf("Full path not found in Implementation imports for action %q", actionRef)
}

func NewActionReferencePatternError(actionRef string) error {
	return errors.Errorf("Action reference %q doesn't follow pattern <import_alias>.<method_name>", actionRef)
}

func ResolveActionPathFromImports(imports []*ImplementationImport, actionRef string) (*InterfaceReference, error) {
	action := strings.SplitN(actionRef, ".", 2)
	if len(action) != 2 {
		return nil, NewActionReferencePatternError(actionRef)
	}

	alias, name := action[0], action[1]
	selectFirstMatchedImport := func() *InterfaceReference {
		for _, i := range imports {
			if i.Alias == nil || *i.Alias != alias {
				continue
			}
			for _, method := range i.Methods {
				if name != method.Name {
					continue
				}
				return &InterfaceReference{
					Path:     fmt.Sprintf("%s.%s", i.InterfaceGroupPath, name),
					Revision: ptr.StringPtrToString(method.Revision),
				}
			}
		}
		return nil
	}

	ref := selectFirstMatchedImport()
	if ref == nil {
		return nil, NewActionImportsError(actionRef)
	}

	return ref, nil
}
