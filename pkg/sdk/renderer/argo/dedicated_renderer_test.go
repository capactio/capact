package argo

import (
	"testing"

	hubpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"gotest.tools/assert"
)

func TestResolveActionFromImports(t *testing.T) {
	d := dedicatedRenderer{}
	name := "helm"
	appVersion := "3.x.x"
	revision1 := "0.1.0"

	tests := []struct {
		name       string
		shouldFail bool

		imports   []*hubpublicapi.ImplementationImport
		actionRef string

		reference hubpublicapi.InterfaceReference
	}{
		{
			name:       "missing imports",
			shouldFail: true,
			imports:    []*hubpublicapi.ImplementationImport{},
			actionRef:  "helm.install",
			reference:  hubpublicapi.InterfaceReference{},
		},
		{
			name: "correct revision",
			imports: []*hubpublicapi.ImplementationImport{
				{
					InterfaceGroupPath: "cap.interface.runner.helm",
					Alias:              &name,
					AppVersion:         &appVersion,
					Methods: []*hubpublicapi.ImplementationImportMethod{
						{
							Name:     "install",
							Revision: &revision1,
						},
					},
				},
			},
			actionRef: "helm.install",
			reference: hubpublicapi.InterfaceReference{
				Path:     "cap.interface.runner.helm.install",
				Revision: revision1,
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			reference, err := d.resolveActionPathFromImports(test.imports, test.actionRef)
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
