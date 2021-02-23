package argo

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/google/uuid"

	ochlocalgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/local"

	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
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
	genUUID         func() string
}

func NewTypeInstanceHandler(ochCli OCHClient, ochActionsImage string) *TypeInstanceHandler {
	return &TypeInstanceHandler{
		ochCli:          ochCli,
		ochActionsImage: ochActionsImage,
		genUUID: func() string {
			return uuid.New().String()
		},
	}
}

func (r *TypeInstanceHandler) AddInputTypeInstances(rootWorkflow *Workflow, instances []types.InputTypeInstanceRef) error {
	if len(instances) == 0 {
		return nil
	}

	idx, err := getEntrypointWorkflowIndex(rootWorkflow)
	if err != nil {
		return err
	}

	artifacts := wfv1.Artifacts{}
	var typeInstanceToDownload []string

	for _, tiInput := range instances {
		writePath := path.Join("/", tiInput.Name)
		writePath = writePath + ".yaml"
		artifacts = append(artifacts, wfv1.Artifact{
			Name:       tiInput.Name,
			GlobalName: tiInput.Name,
			Path:       writePath,
		})
		typeInstanceToDownload = append(typeInstanceToDownload, fmt.Sprintf("{%s,%s}", tiInput.ID, writePath))
	}

	template := &wfv1.Template{
		Name: fmt.Sprintf("inject-input-type-instances-%s", r.genUUID()),
		Container: &apiv1.Container{
			Image: r.ochActionsImage,
			Env: []apiv1.EnvVar{
				{
					Name:  "APP_ACTION",
					Value: "DownloadAction",
				},
				{
					Name:  "APP_DOWNLOAD_CONFIG",
					Value: strings.Join(typeInstanceToDownload, ","),
				},
			},
		},
		Outputs: wfv1.Outputs{
			Artifacts: artifacts,
		},
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

func (r *TypeInstanceHandler) AddUploadTypeInstancesStep(rootWorkflow *Workflow, output *OutputTypeInstances) error {
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
			Image:           r.ochActionsImage,
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

func (r *TypeInstanceHandler) SetGenUUID(genUUID func() string) {
	r.genUUID = genUUID
}
