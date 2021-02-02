package argo

import (
	"context"
	"log"
	"testing"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
	"gotest.tools/golden"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"projectvoltron.dev/voltron/pkg/och/client/fake"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"

	"github.com/stretchr/testify/require"
)

// TestRenderHappyPath tests that renderer generates valid Argo Workflows.
//
// This test is based on golden file.
// If the `-test.update-golden` flag is set then the actual content is written
// to the golden file.
//
// To update golden file, run:
//   go test ./pkg/sdk/renderer/argo/...  -v -test.update-golden
func TestRenderHappyPath(t *testing.T) {
	// given
	fakeCli, err := fake.NewFromLocal("testdata/och")
	require.NoError(t, err)

	renderer := NewRenderer(fakeCli)

	tests := []struct {
		name               string
		ref                types.InterfaceRef
		userInput          map[string]interface{}
		inputTypeInstances []types.InputTypeInstanceRef
	}{
		{
			name: "PostgreSQL workflow without user input and TypeInstances",
			ref: types.InterfaceRef{
				Path: "cap.interface.database.postgresql.install",
			},
		},
		{
			name: "PostgreSQL workflow with user input and without TypeInstances",
			ref: types.InterfaceRef{
				Path: "cap.interface.database.postgresql.install",
			},
			userInput: map[string]interface{}{
				"superuser": map[string]interface{}{
					"username": "postgres",
					"password": "s3cr3t",
				},
				"defaultDBName": "test",
			},
		},
		{
			name: "Jira workflow with user input and TypeInstances",
			ref: types.InterfaceRef{
				Path: "cap.interface.productivity.jira.install",
			},
			userInput: map[string]interface{}{
				"superuser": map[string]interface{}{
					"username": "postgres",
					"password": "s3cr3t",
				},
				"defaultDBName": "test",
			},

			inputTypeInstances: []types.InputTypeInstanceRef{
				{
					Name: "gcp",
					ID:   "c268d3f5-8834-434b-bea2-b677793611c5",
				},
			},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			renderedArgs, err := renderer.Render(context.Background(), tt.ref, WithPlainTextUserInput(tt.userInput), WithTypeInstances(tt.inputTypeInstances))
			require.NoError(t, err)

			// then
			AssertYAMLGoldenFile(t, renderedArgs, t.Name())
			saveToRun(t, renderedArgs)
		})
	}
}

func AssertYAMLGoldenFile(t *testing.T, actualYAMLData interface{}, filename string, msgAndArgs ...interface{}) {
	t.Helper()

	out, err := yaml.Marshal(actualYAMLData)
	require.NoError(t, err)
	golden.Assert(t, string(out), filename+".golden.yaml", msgAndArgs)
}

// TODO(mszostok): Remove
func saveToRun(t *testing.T, action *types.Action) {
	obj := &unstructured.Unstructured{}

	obj.SetKind("Workflow")
	obj.SetAPIVersion("argoproj.io/v1alpha1")
	obj.SetName("workflow-my")

	if err := mapstructure.Decode(map[string]interface{}{
		"spec": action.Args["workflow"],
	}, &obj.Object); err != nil {
		log.Fatal(err)
	}

	yamlData, err := yaml.Marshal(obj.Object)
	if err != nil {
		log.Fatal(err)
	}

	golden.Assert(t, string(yamlData), t.Name()+".to-run.yaml")
}
