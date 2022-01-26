package manifest

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	hubpublicgraphql "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/hub/client/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
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
			validator := NewRemoteImplementationValidator(hubCli)

			implementation := &types.Implementation{}
			b, err := json.Marshal(tc.argoArtifacts)
			require.NoError(t, err)
			implementationRaw := getImplementationRaw(b)
			require.NoError(t, json.Unmarshal(implementationRaw, implementation))
			implementation.Spec.AdditionalInput = tc.implAdditionalInput

			// when
			result, err := validator.validateInputArtifactsNames(ctx, *implementation)

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.exprectedResults, result)
		})
	}
}

func getImplementationRaw(argoArtifactsInput []byte) []byte {
	return []byte(fmt.Sprintf(`
	{
	  "spec": {
		"appVersion": "1.0.1",
		"implements": [
		  {
			"path": "cap.interface.test.impl",
			"revision": "0.1.0"
		  }
		],
		"action": {
		  "runnerInterface": "argo.run",
		  "args": {
			"workflow": {
			  "entrypoint": "test",
			  "templates": [
				{
				  "name": "test",
				  "inputs": {
					"artifacts": %s,
					"outputs": {
					  "artifacts": []
					}
				  },
				  "steps": []
				}
			  ]
			}
		  }
		}
	  }
	}`, string(argoArtifactsInput)))
}

type fakeHubCli struct {
	InterfaceRevision *hubpublicgraphql.InterfaceRevision
}

func (f fakeHubCli) CheckManifestRevisionsExist(ctx context.Context, manifestRefs []hubpublicgraphql.ManifestReference) (map[hubpublicgraphql.ManifestReference]bool, error) {
	return map[hubpublicgraphql.ManifestReference]bool{}, nil
}
func (f fakeHubCli) FindInterfaceRevision(ctx context.Context, ref hubpublicgraphql.InterfaceReference, opts ...public.InterfaceRevisionOption) (*hubpublicgraphql.InterfaceRevision, error) {
	return f.InterfaceRevision, nil
}
