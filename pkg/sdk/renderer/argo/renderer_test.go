package argo

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/och/client/fake"

	"capact.io/capact/pkg/och/client"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/renderer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gotest.tools/golden"
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
	fakeCli, err := fake.NewFromLocal("testdata/och", false)
	require.NoError(t, err)

	policy := policy.NewAllowAll()
	policyEnforcedCli := client.NewPolicyEnforcedClient(fakeCli)
	genUUID := func() string { return "uuid" } // it has to be static because of parallel testing
	typeInstanceHandler := NewTypeInstanceHandler("alpine:3.7")
	typeInstanceHandler.SetGenUUID(genUUID)

	ownerID := "default/action"

	argoRenderer := NewRenderer(renderer.Config{
		RenderTimeout: time.Second,
		MaxDepth:      20,
	}, policyEnforcedCli, typeInstanceHandler)

	tests := []struct {
		name                string
		ref                 types.InterfaceRef
		inputTypeInstances  []types.InputTypeInstanceRef
		userInput           *UserInputSecretRef
		typeInstancesToLock []string
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
			name: "Mattermost workflow with user input and gcp TypeInstance",
			ref: types.InterfaceRef{
				Path: "cap.interface.productivity.mattermost.install",
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
			name: "Mattermost workflow with user input and gcp and postgresql TypeInstances",
			ref: types.InterfaceRef{
				Path: "cap.interface.productivity.mattermost.install",
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
			name: "Workflow with apps stack installation without user input and TypeInstances",
			ref: types.InterfaceRef{
				Path: "cap.interface.app-stack.stack.install",
			},
		},
		{
			name: "PostgreSQL change password",
			ref: types.InterfaceRef{
				Path: "cap.interface.database.postgresql.change-password",
			},
			inputTypeInstances: []types.InputTypeInstanceRef{
				{
					Name: "role",
					ID:   "6fc7dd6b-d150-4af3-a1aa-a868962b7d68",
				},
				{
					Name: "postgresql",
					ID:   "f2421415-b8a4-464b-be12-b617794411c5",
				},
			},
			userInput: &UserInputSecretRef{
				Name: "user-input",
				Key:  "parameters.json",
			},
			typeInstancesToLock: []string{"6fc7dd6b-d150-4af3-a1aa-a868962b7d68"},
		},
		{
			name: "Nested PostgreSQL change password",
			ref: types.InterfaceRef{
				Path: "cap.interface.nested.change-password",
			},
			inputTypeInstances: []types.InputTypeInstanceRef{
				{
					Name: "firstRole",
					ID:   "6fc7dd6b-d150-4af3-a1aa-a868962b7d68",
				},
				{
					Name: "postgresql",
					ID:   "f2421415-b8a4-464b-be12-b617794411c5",
				},
			},
			userInput: &UserInputSecretRef{
				Name: "user-input",
				Key:  "parameters.json",
			},
			typeInstancesToLock: []string{"6fc7dd6b-d150-4af3-a1aa-a868962b7d68"},
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
			renderOutput, err := argoRenderer.Render(
				context.Background(),
				&RenderInput{
					RunnerContextSecretRef: RunnerContextSecretRef{Name: "secret", Key: "key"},
					InterfaceRef:           tt.ref,
					Options: []RendererOption{
						WithSecretUserInput(tt.userInput),
						WithTypeInstances(tt.inputTypeInstances),
						WithPolicy(policy),
						WithOwnerID(ownerID),
					},
				},
			)

			// then
			require.NoError(t, err)
			assertYAMLGoldenFile(t, renderOutput.Action, t.Name())
			assert.Equal(t, tt.typeInstancesToLock, renderOutput.TypeInstancesToLock)
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
	fakeCli, err := fake.NewFromLocal("testdata/och", true)
	require.NoError(t, err)

	tests := []struct {
		name               string
		ref                types.InterfaceRef
		inputTypeInstances []types.InputTypeInstanceRef
		policy             policy.Policy
	}{
		{
			name: "Mattermost with CloudSQL PostgreSQL installation with GCP SA injected",
			ref: types.InterfaceRef{
				Path: "cap.interface.productivity.mattermost.install",
			},
			inputTypeInstances: []types.InputTypeInstanceRef{
				{
					Name: "foo",
					ID:   "c268d3f5-8834-434b-bea2-b677793611c5",
				},
			},
			policy: fixGCPGlobalPolicy(),
		},
		{
			name: "CloudSQL PostgreSQL installation with GCP SA injected",
			ref: types.InterfaceRef{
				Path: "cap.interface.database.postgresql.install",
			},
			policy: fixGCPGlobalPolicy(),
		},
		{
			name: "Mattermost with CloudSQL using Terraform",
			ref: types.InterfaceRef{
				Path: "cap.interface.productivity.mattermost.install",
			},
			policy: fixTerraformPolicy(),
		},
		{
			name: "Mattermost with AWS RDS install",
			ref: types.InterfaceRef{
				Path: "cap.interface.productivity.mattermost.install",
			},
			policy: fixAWSGlobalPolicy(),
		},
		{
			name: "Unmet policy constraints - fallback to Bitnami Implementation",
			ref: types.InterfaceRef{
				Path: "cap.interface.database.postgresql.install",
			},
			policy: fixGlobalPolicyForFallback(),
		},
	}
	for testIdx, test := range tests {
		tc := testIdx
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			policyEnforcedCli := client.NewPolicyEnforcedClient(fakeCli)
			genUUID := genUUIDFn(strconv.Itoa(tc))
			typeInstanceHandler := NewTypeInstanceHandler("alpine:3.7")
			typeInstanceHandler.SetGenUUID(genUUID)

			argoRenderer := NewRenderer(renderer.Config{
				RenderTimeout: time.Hour,
				MaxDepth:      50,
			}, policyEnforcedCli, typeInstanceHandler)

			// when
			renderOutput, err := argoRenderer.Render(
				context.Background(),
				&RenderInput{
					RunnerContextSecretRef: RunnerContextSecretRef{Name: "secret", Key: "key"},
					InterfaceRef:           tt.ref,
					Options: []RendererOption{
						WithTypeInstances(tt.inputTypeInstances),
						WithPolicy(tt.policy),
						WithOwnerID("owner"),
					},
				},
			)

			// then
			require.NoError(t, err)
			assertYAMLGoldenFile(t, renderOutput.Action, t.Name())
		})
	}
}

func TestRendererMaxDepth(t *testing.T) {
	// given
	fakeCli, err := fake.NewFromLocal("testdata/och", false)
	require.NoError(t, err)

	policy := policy.NewAllowAll()
	policyEnforcedCli := client.NewPolicyEnforcedClient(fakeCli)
	typeInstanceHandler := NewTypeInstanceHandler("alpine:3.7")
	typeInstanceHandler.SetGenUUID(genUUIDFn(""))

	argoRenderer := NewRenderer(renderer.Config{
		RenderTimeout: time.Second,
		MaxDepth:      3,
	}, policyEnforcedCli, typeInstanceHandler)

	interfaceRef := types.InterfaceRef{
		Path: "cap.interface.infinite.render.loop",
	}

	// when
	renderOutput, err := argoRenderer.Render(
		context.Background(),
		&RenderInput{
			RunnerContextSecretRef: RunnerContextSecretRef{Name: "secret", Key: "key"},
			InterfaceRef:           interfaceRef,
			Options: []RendererOption{
				WithPolicy(policy),
			},
		},
	)

	// then
	assert.EqualError(t, err, "Exceeded maximum render depth level [max depth 3]")
	assert.Nil(t, renderOutput)
}

func TestRendererDenyAllPolicy(t *testing.T) {
	// given
	fakeCli, err := fake.NewFromLocal("testdata/och", false)
	require.NoError(t, err)

	policy := policy.NewDenyAll()
	policyEnforcedCli := client.NewPolicyEnforcedClient(fakeCli)
	typeInstanceHandler := NewTypeInstanceHandler("alpine:3.7")
	typeInstanceHandler.SetGenUUID(genUUIDFn(""))

	argoRenderer := NewRenderer(renderer.Config{
		RenderTimeout: time.Second,
		MaxDepth:      3,
	}, policyEnforcedCli, typeInstanceHandler)

	interfaceRef := types.InterfaceRef{
		Path: "cap.interface.productivity.mattermost.install",
	}

	// when
	renderOutput, err := argoRenderer.Render(
		context.Background(),
		&RenderInput{
			RunnerContextSecretRef: RunnerContextSecretRef{Name: "secret", Key: "key"},
			InterfaceRef:           interfaceRef,
			Options: []RendererOption{
				WithPolicy(policy),
			},
		},
	)

	// then
	assert.EqualError(t, err,
		`while picking ImplementationRevision for Interface "cap.interface.productivity.mattermost.install:": No Implementations found with current policy for given Interface`)
	assert.Nil(t, renderOutput)
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
