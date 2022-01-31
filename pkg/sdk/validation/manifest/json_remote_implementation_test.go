package manifest

import (
	"context"
	"fmt"
	"testing"

	hubpublicgraphql "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/hub/client/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/renderer/argo"
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateInputArtifactsNames(t *testing.T) {
	inputParametersName := "input-parameters"
	additionalParameterName := "additional-parameters"
	tests := map[string]struct {
		ifaceInput          *hubpublicgraphql.InterfaceInput
		implAdditionalInput *types.AdditionalInput
		argoArtifacts       []wfv1.Artifact
		exprectedResults    ValidationResult
	}{
		"When Argo input does not exist in the Interface and Implementation": {
			ifaceInput: &hubpublicgraphql.InterfaceInput{
				Parameters: []*hubpublicgraphql.InputParameter{},
			},
			implAdditionalInput: &types.AdditionalInput{
				Parameters: map[string]types.AdditionalInputParameter{
					additionalParameterName: {},
				},
			},
			argoArtifacts: []wfv1.Artifact{
				{
					Name: inputParametersName,
				},
				{
					Name:     additionalParameterName,
					Optional: true,
				},
			},
			exprectedResults: newValidationResult(fmt.Errorf("unknown workflow input artifact \"%s\": there is no such input neither in Interface input, nor Implementation additional input", inputParametersName)),
		},
		"When Argo input has optional input that does exist in Interface ": {
			ifaceInput: &hubpublicgraphql.InterfaceInput{
				Parameters: []*hubpublicgraphql.InputParameter{
					{
						Name: inputParametersName,
					},
				},
			},
			implAdditionalInput: &types.AdditionalInput{},
			argoArtifacts: []wfv1.Artifact{
				{
					Name:     inputParametersName,
					Optional: true,
				},
			},
			exprectedResults: newValidationResult(fmt.Errorf("invalid workflow input artifact \"%s\": it shouldn't be optional as it is defined as Interface input", inputParametersName)),
		},
		"When Argo input is not optional but exists in Implementation additional inputs": {
			ifaceInput: &hubpublicgraphql.InterfaceInput{
				Parameters: []*hubpublicgraphql.InputParameter{
					{
						Name: inputParametersName,
					},
				},
			},
			implAdditionalInput: &types.AdditionalInput{
				Parameters: map[string]types.AdditionalInputParameter{
					additionalParameterName: {},
				},
			},
			argoArtifacts: []wfv1.Artifact{
				{
					Name:     additionalParameterName,
					Optional: false,
				},
			},
			exprectedResults: newValidationResult(fmt.Errorf("invalid workflow input artifact \"%s\": it should be optional, as it is defined as Implementation additional input", additionalParameterName)),
		},
		"When Argo inputs are correctly set": {
			ifaceInput: &hubpublicgraphql.InterfaceInput{
				Parameters: []*hubpublicgraphql.InputParameter{
					{
						Name: inputParametersName,
					},
				},
			},
			implAdditionalInput: &types.AdditionalInput{
				Parameters: map[string]types.AdditionalInputParameter{
					additionalParameterName: {},
				},
			},
			argoArtifacts: []wfv1.Artifact{
				{
					Name:     inputParametersName,
					Optional: false,
				},
				{
					Name:     additionalParameterName,
					Optional: true,
				},
			},
			exprectedResults: newValidationResult(),
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			ctx := context.Background()
			hubCli := fakeHubCli{
				InterfaceRevision: &hubpublicgraphql.InterfaceRevision{
					Spec: &hubpublicgraphql.InterfaceSpec{
						Input: tc.ifaceInput,
					},
				},
			}
			validator := NewRemoteImplementationValidator(&hubCli)
			implementation := fixImplementation(tc.argoArtifacts, tc.implAdditionalInput)

			// when
			result, err := validator.validateInputArtifactsNames(ctx, implementation)

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.exprectedResults, result)
		})
	}
}

func fixImplementation(inputArtifacts []wfv1.Artifact, implAdditionalInput *types.AdditionalInput) types.Implementation {
	workflow := argo.Workflow{
		WorkflowSpec: &wfv1.WorkflowSpec{
			Entrypoint: "test",
		},
		Templates: []*argo.Template{
			{
				Template: &wfv1.Template{
					Name: "test",
					Inputs: wfv1.Inputs{
						Artifacts: inputArtifacts,
					},
				},
			},
		},
	}

	return types.Implementation{
		Spec: types.ImplementationSpec{
			Implements: []types.Implement{
				{
					Path:     "cap.interface.test.impl",
					Revision: "0.1.0",
				},
			},
			Action: types.Action{
				Args: map[string]interface{}{
					"workflow": workflow,
				},
			},
			AdditionalInput: implAdditionalInput,
		},
	}
}

type fakeHubCli struct {
	InterfaceRevision *hubpublicgraphql.InterfaceRevision
}

func (f *fakeHubCli) CheckManifestRevisionsExist(_ context.Context, _ []hubpublicgraphql.ManifestReference) (map[hubpublicgraphql.ManifestReference]bool, error) {
	return map[hubpublicgraphql.ManifestReference]bool{}, nil
}
func (f *fakeHubCli) FindInterfaceRevision(_ context.Context, _ hubpublicgraphql.InterfaceReference, _ ...public.InterfaceRevisionOption) (*hubpublicgraphql.InterfaceRevision, error) {
	return f.InterfaceRevision, nil
}
