package graphql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveActionFromImports(t *testing.T) {
	name := "helm"
	appVersion := "3.x.x"
	revision1 := "0.1.0"

	tests := []struct {
		name       string
		shouldFail bool

		imports   []*ImplementationImport
		actionRef string

		reference InterfaceReference
	}{
		{
			name:       "missing imports",
			shouldFail: true,
			imports:    []*ImplementationImport{},
			actionRef:  "helm.install",
			reference:  InterfaceReference{},
		},
		{
			name: "correct revision",
			imports: []*ImplementationImport{
				{
					InterfaceGroupPath: "cap.interface.runner.helm",
					Alias:              &name,
					AppVersion:         &appVersion,
					Methods: []*ImplementationImportMethod{
						{
							Name:     "install",
							Revision: &revision1,
						},
					},
				},
			},
			actionRef: "helm.install",
			reference: InterfaceReference{
				Path:     "cap.interface.runner.helm.install",
				Revision: revision1,
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			reference, err := ResolveActionPathFromImports(test.imports, test.actionRef)
			if test.shouldFail {
				if err == nil {
					t.Fatal("test should fail, but did not")
				}
			} else {
				if err != nil {
					t.Fatalf("test retuned error %v", err)
				}
				assert.Equal(t, test.reference, *reference)
			}
		})
	}
}
