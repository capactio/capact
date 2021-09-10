package argo

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	actionvalidation "capact.io/capact/pkg/sdk/validation/interfaceio"
	policyvalidation "capact.io/capact/pkg/sdk/validation/policy"

	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/hub/client/fake"
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
	fakeCli, err := fake.NewFromLocal("testdata/hub", true)
	require.NoError(t, err)

	policy := policy.NewAllowAll()
	genUUID := func() string { return "uuid" } // it has to be static because of parallel testing
	typeInstanceHandler := NewTypeInstanceHandler("alpine:3.7")
	typeInstanceHandler.SetGenUUID(genUUID)

	ownerID := "default/action"

	interfaceIOValidator := actionvalidation.NewValidator(fakeCli)
	policyIOValidator := policyvalidation.NewValidator(fakeCli)
	wfValidator := renderer.NewWorkflowInputValidator(interfaceIOValidator, policyIOValidator)

	argoRenderer := NewRenderer(renderer.Config{
		RenderTimeout: time.Second,
		MaxDepth:      20,
	}, fakeCli, typeInstanceHandler, wfValidator)

	tests := []struct {
		name                    string
		ref                     types.InterfaceRef
		inputTypeInstances      []types.InputTypeInstanceRef
		userParameterCollection types.ParametersCollection
		typeInstancesToLock     []string
	}{
		{
			name: "Two level nested workflow without user input and TypeInstances",
			ref: types.InterfaceRef{
				Path: "cap.interface.nested.root",
			},
		},
		{
			name: "PostgreSQL workflow with user input and without TypeInstances",
			ref: types.InterfaceRef{
				Path: "cap.interface.database.postgresql.install",
			},
			userParameterCollection: types.ParametersCollection{
				"input-parameters": `{"superuser":{"password":"bar"}}}`,
			},
		},
		{
			name: "Workflow with apps stack installation with user input",
			ref: types.InterfaceRef{
				Path: "cap.interface.app-stack.stack.install",
			},
			userParameterCollection: types.ParametersCollection{
				"input-parameters": `{"key":true}`,
			},
		},
		{
			name: "Mattermost workflow with user input",
			ref: types.InterfaceRef{
				Path: "cap.interface.productivity.mattermost.install",
			},
			userParameterCollection: types.ParametersCollection{
				"input-parameters": `{"host":"mattermost.local"}`,
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
			userParameterCollection: types.ParametersCollection{
				"input-parameters": `{"password":"foo"}`,
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
			userParameterCollection: types.ParametersCollection{
				"input-parameters": `{"key":true}`,
			},
			typeInstancesToLock: []string{"6fc7dd6b-d150-4af3-a1aa-a868962b7d68"},
		},
		{
			name: "Workflow with two input parameters",
			ref: types.InterfaceRef{
				Path: "cap.interface.multiparam.two",
			},
			userParameterCollection: types.ParametersCollection{
				"input-parameters":  `{"key":true}`,
				"second-parameters": `{"key":false}`,
			},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			opts := []RendererOption{
				WithTypeInstances(tt.inputTypeInstances),
				WithGlobalPolicy(policy),
				WithOwnerID(ownerID),
			}

			if len(tt.userParameterCollection) > 0 {
				opts = append(opts, WithSecretUserInput(&UserInputSecretRef{
					Name: "user-input",
				}, tt.userParameterCollection))
			}

			// when
			renderOutput, err := argoRenderer.Render(
				context.Background(),
				&RenderInput{
					RunnerContextSecretRef: RunnerContextSecretRef{Name: "secret", Key: "key"},
					InterfaceRef:           tt.ref,
					Options:                opts,
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
	fakeCli, err := fake.NewFromLocal("testdata/hub", true)
	require.NoError(t, err)

	tests := []struct {
		name                    string
		ref                     types.InterfaceRef
		inputTypeInstances      []types.InputTypeInstanceRef
		userParameterCollection types.ParametersCollection
		policy                  policy.Policy
	}{
		{
			name: "Mattermost with CloudSQL PostgreSQL installation with GCP SA injected",
			ref: types.InterfaceRef{
				Path: "cap.interface.productivity.mattermost.install",
			},
			userParameterCollection: types.ParametersCollection{
				"input-parameters": `{"host":"mattermost.local"}`,
			},
			policy: fixGCPGlobalPolicy(),
		},
		{
			name: "Mattermost with existing DB installation",
			ref: types.InterfaceRef{
				Path: "cap.interface.productivity.mattermost.install",
			},
			userParameterCollection: types.ParametersCollection{
				"input-parameters": `{"host":"mattermost.local"}`,
			},
			policy: fixExistingDBPolicy(),
		},
		{
			name: "CloudSQL PostgreSQL installation with GCP SA injected",
			ref: types.InterfaceRef{
				Path: "cap.interface.database.postgresql.install",
			},
			policy: fixGCPGlobalPolicy(),
			userParameterCollection: types.ParametersCollection{
				"input-parameters": `{"superuser":{"password":"bar"}}`,
			},
		},
		{
			name: "RDS installation with AWS SA and additional parameters injected",
			ref: types.InterfaceRef{
				Path: "cap.interface.database.postgresql.install",
			},
			policy: fixAWSRDSPolicy(),
			userParameterCollection: types.ParametersCollection{
				"input-parameters": `{"superuser":{"password":"bar"}}`,
			},
		},
		{
			name: "Mattermost with CloudSQL using Terraform",
			ref: types.InterfaceRef{
				Path: "cap.interface.productivity.mattermost.install",
			},
			policy: fixTerraformPolicy(),
			userParameterCollection: types.ParametersCollection{
				"input-parameters": `{"host":"mattermost.local"}`,
			},
		},
		{
			name: "Mattermost with AWS RDS install",
			ref: types.InterfaceRef{
				Path: "cap.interface.productivity.mattermost.install",
			},
			policy: fixAWSGlobalPolicy(),
			userParameterCollection: types.ParametersCollection{
				"input-parameters": `{"host":"mattermost.local"}`,
			},
		},
		{
			name: "Unmet policy constraints - fallback to Bitnami Implementation",
			ref: types.InterfaceRef{
				Path: "cap.interface.database.postgresql.install",
			},
			policy: fixGlobalPolicyForFallback(),
			userParameterCollection: types.ParametersCollection{
				"input-parameters": `{"superuser":{"password":"bar"}}`,
			},
		},
		{
			name: "Workflow policy injects additional input - reference by ManifestRef",
			ref: types.InterfaceRef{
				Path: "cap.interface.app-stack.app1.install",
			},
			policy: fixAWSGlobalPolicy(),
			userParameterCollection: types.ParametersCollection{
				"input-parameters": `{"key":"string"}`,
			},
		},
		{
			name: "Workflow policy injects additional input - reference by alias",
			ref: types.InterfaceRef{
				Path: "cap.interface.app-stack.app2.install",
			},
			policy: fixAWSGlobalPolicy(),
			userParameterCollection: types.ParametersCollection{
				"input-parameters": `{"key":"string"}`,
			},
		},
	}
	for testIdx, test := range tests {
		tc := testIdx
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			genUUID := genUUIDFn(strconv.Itoa(tc))
			typeInstanceHandler := NewTypeInstanceHandler("alpine:3.7")
			typeInstanceHandler.SetGenUUID(genUUID)

			interfaceIOValidator := actionvalidation.NewValidator(fakeCli)
			policyIOValidator := policyvalidation.NewValidator(fakeCli)
			wfValidator := renderer.NewWorkflowInputValidator(interfaceIOValidator, policyIOValidator)

			argoRenderer := NewRenderer(renderer.Config{
				RenderTimeout: time.Hour,
				MaxDepth:      50,
			}, fakeCli, typeInstanceHandler, wfValidator)

			opts := []RendererOption{
				WithTypeInstances(tt.inputTypeInstances),
				WithGlobalPolicy(tt.policy),
				WithOwnerID("owner"),
			}

			if len(tt.userParameterCollection) > 0 {
				opts = append(opts, WithSecretUserInput(&UserInputSecretRef{
					Name: "user-input",
				}, tt.userParameterCollection))
			}

			// when
			renderOutput, err := argoRenderer.Render(
				context.Background(),
				&RenderInput{
					RunnerContextSecretRef: RunnerContextSecretRef{Name: "secret", Key: "key"},
					InterfaceRef:           tt.ref,
					Options:                opts,
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
	fakeCli, err := fake.NewFromLocal("testdata/hub", false)
	require.NoError(t, err)

	policy := policy.NewAllowAll()
	typeInstanceHandler := NewTypeInstanceHandler("alpine:3.7")
	typeInstanceHandler.SetGenUUID(genUUIDFn(""))

	interfaceIOValidator := actionvalidation.NewValidator(fakeCli)
	policyIOValidator := policyvalidation.NewValidator(fakeCli)
	wfValidator := renderer.NewWorkflowInputValidator(interfaceIOValidator, policyIOValidator)

	argoRenderer := NewRenderer(renderer.Config{
		RenderTimeout: time.Second,
		MaxDepth:      3,
	}, fakeCli, typeInstanceHandler, wfValidator)

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
				WithGlobalPolicy(policy),
			},
		},
	)

	// then
	assert.EqualError(t, err, "Exceeded maximum render depth level [max depth 3]")
	assert.Nil(t, renderOutput)
}

func TestRendererDenyAllPolicy(t *testing.T) {
	// given
	fakeCli, err := fake.NewFromLocal("testdata/hub", false)
	require.NoError(t, err)

	policy := policy.NewDenyAll()
	typeInstanceHandler := NewTypeInstanceHandler("alpine:3.7")
	typeInstanceHandler.SetGenUUID(genUUIDFn(""))

	interfaceIOValidator := actionvalidation.NewValidator(fakeCli)
	policyIOValidator := policyvalidation.NewValidator(fakeCli)
	wfValidator := renderer.NewWorkflowInputValidator(interfaceIOValidator, policyIOValidator)

	argoRenderer := NewRenderer(renderer.Config{
		RenderTimeout: time.Second,
		MaxDepth:      3,
	}, fakeCli, typeInstanceHandler, wfValidator)

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
				WithGlobalPolicy(policy),
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
