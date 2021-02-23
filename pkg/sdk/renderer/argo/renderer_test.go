package argo

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"projectvoltron.dev/voltron/pkg/och/client/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gotest.tools/golden"
	"projectvoltron.dev/voltron/pkg/engine/k8s/clusterpolicy"
	"projectvoltron.dev/voltron/pkg/och/client"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
	"projectvoltron.dev/voltron/pkg/sdk/renderer"
	"sigs.k8s.io/yaml"
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

	policy := clusterpolicy.NewAllowAll()
	policyEnforcedCli := client.NewPolicyEnforcedClient(fakeCli)
	genUUID := func() string { return "uuid" } // it has to be static because of parallel testing
	typeInstanceHandler := NewTypeInstanceHandler(fakeCli, "alpine:3.7", genUUID)

	argoRenderer := NewRenderer(renderer.Config{
		RenderTimeout: time.Second,
		MaxDepth:      20,
	}, policyEnforcedCli, typeInstanceHandler)

	tests := []struct {
		name               string
		ref                types.InterfaceRef
		inputTypeInstances []types.InputTypeInstanceRef
		userInput          *UserInputSecretRef
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
			userInput: &UserInputSecretRef{
				Name: "user-input",
				Key:  "parameters.json",
			},
		},
		{
			name: "Jira workflow with user input and gcp TypeInstance",
			ref: types.InterfaceRef{
				Path: "cap.interface.productivity.jira.install",
			},
			userInput: &UserInputSecretRef{
				Name: "user-input",
				Key:  "parameters.json",
			},
			inputTypeInstances: []types.InputTypeInstanceRef{
				{
					Name: "gcp",
					ID:   "c268d3f5-8834-434b-bea2-b677793611c5",
				},
			},
		},
		{
			name: "Jira workflow with user input and gcp and postgresql TypeInstances",
			ref: types.InterfaceRef{
				Path: "cap.interface.productivity.jira.install",
			},
			userInput: &UserInputSecretRef{
				Name: "user-input",
				Key:  "parameters.json",
			},
			inputTypeInstances: []types.InputTypeInstanceRef{
				{
					Name: "gcp",
					ID:   "c268d3f5-8834-434b-bea2-b677793611c5",
				},
				{
					Name: "postgresql",
					ID:   "f2421415-b8a4-464b-be12-b617794411c5",
				},
			},
		},
		{
			name: "Atlassian stack without user input and TypeInstances",
			ref: types.InterfaceRef{
				Path: "cap.interface.atlassian.stack.install",
			},
		},
		{
			name: "Two level nested workflow",
			ref: types.InterfaceRef{
				Path: "cap.interface.nested.root",
			},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			renderedArgs, err := argoRenderer.Render(
				context.Background(),
				RunnerContextSecretRef{Name: "secret", Key: "key"},
				tt.ref,
				WithSecretUserInput(tt.userInput),
				WithTypeInstances(tt.inputTypeInstances),
				WithPolicy(policy),
			)

			// then
			require.NoError(t, err)
			assertYAMLGoldenFile(t, renderedArgs, t.Name())
		})
	}
}

// TestRenderHappyPathWithCustomPolicies tests that renderer generates valid Argo Workflows with custom policies.
// These test cases are separated from TestRenderHappyPath as they cannot be tested concurrently
// because of different policies set per test case.
//
// This test is based on golden file.
// If the `-test.update-golden` flag is set then the actual content is written
// to the golden file.
//
// To update golden file, run:
//   go test ./pkg/sdk/renderer/argo/...  -v -test.update-golden
func TestRenderHappyPathWithCustomPolicies(t *testing.T) {
	// given
	fakeCli, err := fake.NewFromLocal("testdata/och")
	require.NoError(t, err)

	tests := []struct {
		name               string
		ref                types.InterfaceRef
		inputTypeInstances []types.InputTypeInstanceRef
		policy             clusterpolicy.ClusterPolicy
	}{
		{
			name: "Jira with CloudSQL PostgreSQL installation with GCP SA injected",
			ref: types.InterfaceRef{
				Path: "cap.interface.productivity.jira.install",
			},
			inputTypeInstances: []types.InputTypeInstanceRef{
				{
					Name: "foo",
					ID:   "c268d3f5-8834-434b-bea2-b677793611c5",
				},
			},
			policy: fixGCPClusterPolicy(),
		},
		{
			name: "CloudSQL PostgreSQL installation with GCP SA injected",
			ref: types.InterfaceRef{
				Path: "cap.interface.database.postgresql.install",
			},
			policy: fixGCPClusterPolicy(),
		},
		{
			name: "Unmet policy constraints - fallback to Bitnami Implementation",
			ref: types.InterfaceRef{
				Path: "cap.interface.database.postgresql.install",
			},
			policy: fixClusterPolicyForFallback(),
		},
	}
	for testIdx, test := range tests {
		tc := testIdx
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			policyEnforcedCli := client.NewPolicyEnforcedClient(fakeCli)
			genUUID := genUUIDFn(strconv.Itoa(tc))
			typeInstanceHandler := NewTypeInstanceHandler(fakeCli, "alpine:3.7", genUUID)

			argoRenderer := NewRenderer(renderer.Config{
				RenderTimeout: time.Hour,
				MaxDepth:      20,
			}, policyEnforcedCli, typeInstanceHandler)

			// when
			renderedArgs, err := argoRenderer.Render(
				context.Background(),
				RunnerContextSecretRef{Name: "secret", Key: "key"},
				tt.ref,
				WithTypeInstances(tt.inputTypeInstances),
				WithPolicy(tt.policy),
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

	policy := clusterpolicy.NewAllowAll()
	policyEnforcedCli := client.NewPolicyEnforcedClient(fakeCli)
	typeInstanceHandler := NewTypeInstanceHandler(fakeCli, "alpine:3.7", genUUIDFn(""))

	argoRenderer := NewRenderer(renderer.Config{
		RenderTimeout: time.Second,
		MaxDepth:      3,
	}, policyEnforcedCli, typeInstanceHandler)

	interfaceRef := types.InterfaceRef{
		Path: "cap.interface.infinite.render.loop",
	}

	// when
	renderedArgs, err := argoRenderer.Render(
		context.Background(),
		RunnerContextSecretRef{Name: "secret", Key: "key"},
		interfaceRef,
		WithPolicy(policy))

	// then
	assert.EqualError(t, err, "Exceeded maximum render depth level [max depth 3]")
	assert.Nil(t, renderedArgs)
}

func TestRendererDenyAllPolicy(t *testing.T) {
	// given
	fakeCli, err := fake.NewFromLocal("testdata/och")
	require.NoError(t, err)

	policy := clusterpolicy.NewDenyAll()
	policyEnforcedCli := client.NewPolicyEnforcedClient(fakeCli)
	typeInstanceHandler := NewTypeInstanceHandler(fakeCli, "alpine:3.7", genUUIDFn(""))

	argoRenderer := NewRenderer(renderer.Config{
		RenderTimeout: time.Second,
		MaxDepth:      3,
	}, policyEnforcedCli, typeInstanceHandler)

	interfaceRef := types.InterfaceRef{
		Path: "cap.interface.productivity.jira.install",
	}

	// when
	renderedArgs, err := argoRenderer.Render(
		context.Background(),
		RunnerContextSecretRef{Name: "secret", Key: "key"},
		interfaceRef,
		WithPolicy(policy))

	// then
	assert.EqualError(t, err,
		`while picking ImplementationRevision for Interface "cap.interface.productivity.jira.install:0.1.0": No Implementations found with current policy for given Interface`)
	assert.Nil(t, renderedArgs)
}

func assertYAMLGoldenFile(t *testing.T, actualYAMLData interface{}, filename string, msgAndArgs ...interface{}) {
	t.Helper()

	out, err := yaml.Marshal(actualYAMLData)
	require.NoError(t, err)
	golden.Assert(t, string(out), filename+".golden.yaml", msgAndArgs)
}

func genUUIDFn(prefix string) func() string {
	return func() func() string {
		i := 0
		return func() string {
			uuid := fmt.Sprintf("%s-%s", prefix, strconv.Itoa(i))
			i++
			return uuid
		}
	}()
}
