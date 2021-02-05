package argo

import (
	"context"
	"testing"
	"time"

	"projectvoltron.dev/voltron/pkg/och/client/fake"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
	"projectvoltron.dev/voltron/pkg/sdk/renderer"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gotest.tools/golden"
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

			fakeCli, err := fake.NewFromLocal("testdata/och")
			require.NoError(t, err)

			argoRenderer := NewRenderer(renderer.Config{
				RenderTimeout: time.Second,
				MaxDepth:      10,
			}, fakeCli)

			// when
			renderedArgs, err := argoRenderer.Render(
				context.Background(),
				RunnerContextSecretRef{Name: "secret", Key: "key"},
				tt.ref,
				WithPlainTextUserInput(tt.userInput),
				WithTypeInstances(tt.inputTypeInstances),
			)

			// then
			require.NoError(t, err)
			assertYAMLGoldenFile(t, renderedArgs, t.Name())
		})
	}
}

func TestRendererMaxDepth(t *testing.T) {
	// given
	fakeCli, err := fake.NewFromLocal("testdata/och")
	require.NoError(t, err)

	argoRenderer := NewRenderer(renderer.Config{
		RenderTimeout: time.Second,
		MaxDepth:      3,
	}, fakeCli)

	interfaceRef := types.InterfaceRef{
		Path: "cap.interface.infinite.render.loop",
	}

	// when
	renderedArgs, err := argoRenderer.Render(context.Background(), RunnerContextSecretRef{Name: "secret", Key: "key"}, interfaceRef)

	// then
	assert.EqualError(t, err, "Exceeded maximum render depth level [max depth 3]")
	assert.Nil(t, renderedArgs)
}

func assertYAMLGoldenFile(t *testing.T, actualYAMLData interface{}, filename string, msgAndArgs ...interface{}) {
	t.Helper()

	out, err := yaml.Marshal(actualYAMLData)
	require.NoError(t, err)
	golden.Assert(t, string(out), filename+".golden.yaml", msgAndArgs)
}
