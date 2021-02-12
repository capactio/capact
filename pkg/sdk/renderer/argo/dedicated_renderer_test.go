package argo

import (
	"fmt"
	"testing"

	"gotest.tools/assert"
	ochpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

func TestResolveActionFromImports(t *testing.T) {
	d := dedicatedRenderer{}
	name := "helm"
	appVersion := "3.x.x"
	revision1 := "0.1.0"

	tests := []struct {
		name       string
		shouldFail bool

		imports   []*ochpublicapi.ImplementationImport
		actionRef string

		reference ochpublicapi.InterfaceReference
	}{
		{
			name:       "missing imports",
			shouldFail: true,
			imports:    []*ochpublicapi.ImplementationImport{},
			actionRef:  "helm.run",
			reference:  ochpublicapi.InterfaceReference{},
		},
		{
			name: "correct revision",
			imports: []*ochpublicapi.ImplementationImport{
				{
					InterfaceGroupPath: "cap.interface.runner.helm",
					Alias:              &name,
					AppVersion:         &appVersion,
					Methods: []*ochpublicapi.ImplementationImportMethod{
						{
							Name:     "run",
							Revision: &revision1,
						},
					},
				},
			},
			actionRef: "helm.run",
			reference: ochpublicapi.InterfaceReference{
				Path:     "cap.interface.runner.helm.run",
				Revision: revision1,
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			reference, err := d.resolveActionPathFromImports(test.imports, test.actionRef)
			fmt.Println(err, reference, test.reference)
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