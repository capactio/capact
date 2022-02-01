package manifest_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/renderer/argo"
	"capact.io/capact/pkg/sdk/validation/manifest"
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateInputArtifactsNames(t *testing.T) {
	inputParametersName := "input-parameters"
	additionalParameterName := "additional-parameters"
	tests := map[string]struct {
		ifaceInput          *gqlpublicapi.InterfaceInput
		implAdditionalInput *types.AdditionalInput
		argoArtifacts       []wfv1.Artifact
		exprectedErrors     []error
	}{
		"When Argo input does not exist in the Interface and Implementation": {
			ifaceInput: &gqlpublicapi.InterfaceInput{
				Parameters: []*gqlpublicapi.InputParameter{},
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
			exprectedErrors: []error{fmt.Errorf("unknown workflow input artifact \"%s\": there is no such input neither in Interface input, nor Implementation additional input", inputParametersName)},
		},
		"When Argo input has optional input that does exist in Interface ": {
			ifaceInput: &gqlpublicapi.InterfaceInput{
				Parameters: []*gqlpublicapi.InputParameter{
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
			exprectedErrors: []error{fmt.Errorf("invalid workflow input artifact \"%s\": it shouldn't be optional as it is defined as Interface input", inputParametersName)},
		},
		"When Argo input is not optional but exists in Implementation additional inputs": {
			ifaceInput: &gqlpublicapi.InterfaceInput{
				Parameters: []*gqlpublicapi.InputParameter{
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
			exprectedErrors: []error{fmt.Errorf("invalid workflow input artifact \"%s\": it should be optional, as it is defined as Implementation additional input", additionalParameterName)},
		},
		"When Argo inputs are correctly set": {
			ifaceInput: &gqlpublicapi.InterfaceInput{
				Parameters: []*gqlpublicapi.InputParameter{
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
			exprectedErrors: nil,
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			ctx := context.Background()
			hubCli := fakeHub{
				interfaceRevision: &gqlpublicapi.InterfaceRevision{
					Spec: &gqlpublicapi.InterfaceSpec{
						Input: tc.ifaceInput,
					},
				},
			}
			validator := manifest.NewRemoteImplementationValidator(&hubCli)
			implementation := fixImplementation(tc.argoArtifacts, tc.implAdditionalInput)

			// when
			result, err := validator.ValidateInputArtifactsNames(ctx, implementation)

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.exprectedErrors, result.Errors)
		})
	}
}

func TestCheckParentNodesAssociation(t *testing.T) {
	tests := map[string]struct {
		knownTypes            []*gqlpublicapi.Type
		relationsToParentNode manifest.ParentNodesAssociation
		expErrors             []error
	}{
		"should success as all nodes are attached to parent nodes": {
			knownTypes: []*gqlpublicapi.Type{
				fixGQLType("cap.type.platform.cloud-foundry", "0.1.0", "cap.core.type.platform"),
				fixGQLType("cap.type.platform.nomad", "0.1.0", "cap.core.type.platform"),
			},

			relationsToParentNode: manifest.ParentNodesAssociation{
				"cap.core.type.platform": {
					{Path: "cap.type.platform.cloud-foundry", Revision: "0.1.0"},
					{Path: "cap.type.platform.nomad", Revision: "0.1.0"},
				},
			},

			expErrors: nil, // no errors
		},
		"should detect that both Type with different revision is not attached to parent node": {
			knownTypes: []*gqlpublicapi.Type{
				fixGQLType("cap.type.platform.cloud-foundry", "0.1.0", "cap.core.type.platform"),
				fixGQLType("cap.type.platform.cloud-foundry", "0.2.0", ""),
			},
			relationsToParentNode: manifest.ParentNodesAssociation{
				"cap.core.type.platform": {
					{Path: "cap.type.platform.cloud-foundry", Revision: "0.2.0"},
				},
			},

			expErrors: []error{errors.New(`Type "cap.type.platform.cloud-foundry:0.2.0" is not attached to "cap.core.type.platform" parent node`)},
		},
		"should detect that one Type is not attached to parent node (singular)": {
			knownTypes: []*gqlpublicapi.Type{
				fixGQLType("cap.type.platform.nomad", "0.1.0", "cap.core.type.platform"),
				fixGQLType("cap.type.platform.cloud-foundry", "0.1.0", ""),
			},
			relationsToParentNode: manifest.ParentNodesAssociation{
				"cap.core.type.platform": {
					{Path: "cap.type.platform.cloud-foundry", Revision: "0.1.0"},
					{Path: "cap.type.platform.nomad", Revision: "0.1.0"},
				},
			},

			expErrors: []error{errors.New(`Type "cap.type.platform.cloud-foundry:0.1.0" is not attached to "cap.core.type.platform" parent node`)},
		},
		"should detect that both Types are not attached to parent node (plural)": {
			knownTypes: []*gqlpublicapi.Type{
				fixGQLType("cap.type.platform.nomad", "0.1.0", ""),
				fixGQLType("cap.type.platform.cloud-foundry", "0.1.0", ""),
				fixGQLType("cap.type.platform.mesos", "0.1.0", ""),
			},

			relationsToParentNode: manifest.ParentNodesAssociation{
				"cap.core.type.platform": {
					{Path: "cap.type.platform.cloud-foundry", Revision: "0.1.0"},
					{Path: "cap.type.platform.nomad", Revision: "0.1.0"},
					{Path: "cap.type.platform.mesos", Revision: "0.1.0"},
				},
			},

			expErrors: []error{errors.New(`Types "cap.type.platform.cloud-foundry:0.1.0", "cap.type.platform.nomad:0.1.0" and "cap.type.platform.mesos:0.1.0" are not attached to "cap.core.type.platform" parent node`)},
		},

		"should not report problems with parents for unknown Types": {
			knownTypes: nil, // not Types in Hub

			relationsToParentNode: manifest.ParentNodesAssociation{
				"cap.core.type.platform": {
					{Path: "cap.type.platform.cloud-foundry", Revision: "0.1.0"},
					{Path: "cap.type.platform.nomad", Revision: "0.1.0"},
					{Path: "cap.type.platform.mesos", Revision: "0.1.0"},
				},
			},

			expErrors: nil, // no error about parents.
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			fakeHubCli := &fakeHub{knownTypes: tc.knownTypes}

			implValidator := manifest.NewRemoteImplementationValidator(fakeHubCli)

			// when
			result, err := implValidator.CheckParentNodesAssociation(context.Background(), tc.relationsToParentNode)

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.expErrors, result.Errors)
		})
	}
}

func fixGQLType(path, rev, parent string) *gqlpublicapi.Type {
	return &gqlpublicapi.Type{
		Path: path,
		Revisions: []*gqlpublicapi.TypeRevision{
			{
				Revision: rev,
				Spec: &gqlpublicapi.TypeSpec{
					AdditionalRefs: []string{parent},
				},
			}},
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
