package argo

import (
	"context"
	"fmt"

	ochlocalgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/local"

	"projectvoltron.dev/voltron/pkg/och/client"

	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	graphqllocal "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
	"sigs.k8s.io/yaml"
)

type OCHClient interface {
	GetTypeInstance(ctx context.Context, id string) (*ochlocalgraphql.TypeInstance, error)
}

// TypeInstanceHandler provides functionality to handle TypeInstance operations such as
// injecting download step and upload step.
type TypeInstanceHandler struct {
	ochCli          OCHClient
	ochActionsImage string
}

func NewTypeInstanceHandler(ochCli client.OCHClient, ochActionsImage string) *TypeInstanceHandler {
	return &TypeInstanceHandler{ochCli: ochCli, ochActionsImage: ochActionsImage}
}

// TODO(SV-189): Handle that properly
func (r *TypeInstanceHandler) AddInputTypeInstance(rootWorkflow *Workflow, instances []types.InputTypeInstanceRef) error {
	idx, err := getEntrypointWorkflowIndex(rootWorkflow)
	if err != nil {
		return err
	}

	for _, tiInput := range instances {
		template, err := r.getInjectTypeInstanceTemplate(tiInput)
		if err != nil {
			return errors.Wrapf(err, "while getting inject TypeInstance template for %s", tiInput.ID)
		}

		rootWorkflow.Templates[idx].Steps = append([]ParallelSteps{
			{
				&WorkflowStep{
					WorkflowStep: &wfv1.WorkflowStep{
						Name:     fmt.Sprintf("%s-step", template.Name),
						Template: template.Name,
					},
				},
			},
		}, rootWorkflow.Templates[idx].Steps...)

		rootWorkflow.Templates = append(rootWorkflow.Templates, &Template{Template: template})
	}

	return nil
}

type OutputTypeInstanceRelation struct {
	From *string
	To   *string
}

type OutputTypeInstance struct {
	ArtifactName *string
	TypeInstance types.OutputTypeInstance
}

type OutputTypeInstances struct {
	typeInstances []OutputTypeInstance
	relations     []OutputTypeInstanceRelation
}

func (r *TypeInstanceHandler) AddUploadTypeInstancesTemplate(rootWorkflow *Workflow, output *OutputTypeInstances) error {
	artifacts := wfv1.Artifacts{}
	arguments := wfv1.Artifacts{}

	payload := &graphqllocal.CreateTypeInstancesInput{
		TypeInstances: []*graphqllocal.CreateTypeInstanceInput{},
		UsesRelations: []*graphqllocal.TypeInstanceUsesRelationInput{},
	}

	for _, ti := range output.typeInstances {
		payload.TypeInstances = append(payload.TypeInstances, &graphqllocal.CreateTypeInstanceInput{
			Alias: ti.ArtifactName,
			TypeRef: &graphqllocal.TypeReferenceInput{
				Path:     ti.TypeInstance.TypeRef.Path,
				Revision: *ti.TypeInstance.TypeRef.Revision,
			},
			Attributes: []*graphqllocal.AttributeReferenceInput{},
		})

		artifacts = append(artifacts, wfv1.Artifact{
			Name: *ti.ArtifactName,
			Path: fmt.Sprintf("/upload/typeInstances/%s", *ti.ArtifactName),
		})

		arguments = append(arguments, wfv1.Artifact{
			Name: *ti.ArtifactName,
			From: fmt.Sprintf("{{workflow.outputs.artifacts.%s}}", *ti.ArtifactName),
		})
	}

	for _, relation := range output.relations {
		payload.UsesRelations = append(payload.UsesRelations, &graphqllocal.TypeInstanceUsesRelationInput{
			From: *relation.From,
			To:   *relation.To,
		})
	}

	payloadBytes, _ := yaml.Marshal(payload)

	arguments = append(arguments, wfv1.Artifact{
		Name: "payload",
		ArtifactLocation: wfv1.ArtifactLocation{
			Raw: &wfv1.RawArtifact{
				Data: string(payloadBytes),
			},
		},
	})

	artifacts = append(artifacts, wfv1.Artifact{
		Name: "payload",
		Path: "/upload/payload",
	})

	template := &wfv1.Template{
		Name: "upload-output-type-instances",
		Container: &apiv1.Container{
			Image:           "local/argo-actions:dev-9704",
			ImagePullPolicy: apiv1.PullIfNotPresent,
			Env: []apiv1.EnvVar{
				{
					Name:  "APP_ACTION",
					Value: "UploadAction",
				},
				{
					Name:  "APP_UPLOAD_CONFIG_PAYLOAD_FILEPATH",
					Value: "/upload/payload",
				},
				{
					Name:  "APP_UPLOAD_CONFIG_TYPE_INSTANCES_DIR",
					Value: "/upload/typeInstances",
				},
			},
		},
		Inputs: wfv1.Inputs{
			Artifacts: artifacts,
		},
	}

	idx, err := getEntrypointWorkflowIndex(rootWorkflow)
	if err != nil {
		return err
	}

	rootWorkflow.Templates[idx].Steps = append(rootWorkflow.Templates[idx].Steps, ParallelSteps{
		{
			WorkflowStep: &wfv1.WorkflowStep{
				Name:     fmt.Sprintf("%s-step", template.Name),
				Template: template.Name,
				Arguments: wfv1.Arguments{
					Artifacts: arguments,
				},
			},
		},
	})

	rootWorkflow.Templates = append(rootWorkflow.Templates, &Template{Template: template})

	return nil
}

func (r *TypeInstanceHandler) getInjectTypeInstanceTemplate(input types.InputTypeInstanceRef) (*wfv1.Template, error) {
	// this will be removed in SV-189
	typeInstance, err := r.ochCli.GetTypeInstance(context.TODO(), input.ID)
	if err != nil {
		return nil, err
	}
	if typeInstance == nil {
		return nil, fmt.Errorf("failed to find TypeInstance %s", input.ID)
	}

	data, err := yaml.Marshal(typeInstance.Spec.Value)
	if err != nil {
		return nil, errors.Wrap(err, "while to marshal TypeInstance to YAML")
	}

	return &wfv1.Template{
		Name: fmt.Sprintf("inject-%s", input.Name),
		Container: &apiv1.Container{
			Image:   r.ochActionsImage,
			Command: []string{"sh", "-c"},
			Args:    []string{fmt.Sprintf("sleep 2 && echo '%s' | tee /output", string(data))},
		},
		Outputs: wfv1.Outputs{
			Artifacts: wfv1.Artifacts{
				{
					Name:       input.Name,
					GlobalName: input.Name,
					Path:       "/output",
				},
			},
		},
	}, nil
}
