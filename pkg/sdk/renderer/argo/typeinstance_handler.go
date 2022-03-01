package argo

import (
	"fmt"
	"path"
	"strings"

	"capact.io/capact/pkg/engine/k8s/policy"
	graphqllocal "capact.io/capact/pkg/hub/api/graphql/local"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/google/uuid"
	apiv1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// TypeInstanceHandler provides functionality to handle TypeInstance operations such as
// injecting download step and upload step.
type TypeInstanceHandler struct {
	hubActionsImage   string
	localHubEndpoint  string
	publicHubEndpoint string
	genUUID           func() string
}

// NewTypeInstanceHandler returns a new TypeInstanceHandler instance.
func NewTypeInstanceHandler(hubActionsImage string, localHubEndpoint string, publicHubEndpoint string) *TypeInstanceHandler {
	return &TypeInstanceHandler{
		hubActionsImage:   hubActionsImage,
		localHubEndpoint:  localHubEndpoint,
		publicHubEndpoint: publicHubEndpoint,
		genUUID: func() string {
			return uuid.New().String()
		},
	}
}

// AddInputTypeInstances adds steps to the workflow to download
// the input TypeInstances.
func (r *TypeInstanceHandler) AddInputTypeInstances(rootWorkflow *Workflow, instances []types.InputTypeInstanceRef) error {
	if len(instances) == 0 {
		return nil
	}

	idx, err := GetEntrypointWorkflowIndex(rootWorkflow)
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
			Image: r.hubActionsImage,
			Env: []apiv1.EnvVar{
				{
					Name:  "APP_ACTION",
					Value: "DownloadAction",
				},
				{
					Name:  "APP_DOWNLOAD_CONFIG",
					Value: strings.Join(typeInstanceToDownload, ","),
				},
				{
					Name:  "APP_LOCAL_HUB_ENDPOINT",
					Value: r.localHubEndpoint,
				},
				{
					Name:  "APP_PUBLIC_HUB_ENDPOINT",
					Value: r.publicHubEndpoint,
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

// OutputTypeInstanceRelation holds a relationship between the output TypeInstances.
type OutputTypeInstanceRelation struct {
	From *string
	To   *string
}

// OutputTypeInstance holds details about a output TypeInstance,
// which will be created in the workflow.
type OutputTypeInstance struct {
	ArtifactName *string
	TypeInstance types.OutputTypeInstance
	Backend      policy.TypeInstanceBackend
}

// OutputTypeInstances holds information about the output TypeInstances
// created in the workflow and the relationships between them.
type OutputTypeInstances struct {
	typeInstances []OutputTypeInstance
	relations     []OutputTypeInstanceRelation
}

// UpdateTypeInstance holds details about a TypeInstance,
// which will be updated in the workflow.
type UpdateTypeInstance struct {
	ArtifactName string
	ID           string
}

// UpdateTypeInstances holds the TypeInstances,
// which will be updates in the workflow.
type UpdateTypeInstances []UpdateTypeInstance

// AddUploadTypeInstancesStep adds workflow steps to upload TypeInstances to the Capact Local Hub.
func (r *TypeInstanceHandler) AddUploadTypeInstancesStep(rootWorkflow *Workflow, output *OutputTypeInstances, ownerID string) error {
	artifacts := wfv1.Artifacts{}
	arguments := wfv1.Artifacts{}

	payload := &graphqllocal.CreateTypeInstancesInput{
		TypeInstances: []*graphqllocal.CreateTypeInstanceInput{},
		UsesRelations: []*graphqllocal.TypeInstanceUsesRelationInput{},
	}

	for _, ti := range output.typeInstances {
		gqlTI := &graphqllocal.CreateTypeInstanceInput{
			Alias:     ti.ArtifactName,
			CreatedBy: &ownerID,
			TypeRef: &graphqllocal.TypeInstanceTypeReferenceInput{
				Path:     ti.TypeInstance.TypeRef.Path,
				Revision: ti.TypeInstance.TypeRef.Revision,
			},
			Attributes: []*graphqllocal.AttributeReferenceInput{},
		}
		if ti.Backend.ID != "" {
			gqlTI.Backend = &graphqllocal.TypeInstanceBackendInput{
				ID: ti.Backend.ID,
			}
		}
		payload.TypeInstances = append(payload.TypeInstances, gqlTI)

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
			Image:           r.hubActionsImage,
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
				{
					Name:  "APP_LOCAL_HUB_ENDPOINT",
					Value: r.localHubEndpoint,
				},
				{
					Name:  "APP_PUBLIC_HUB_ENDPOINT",
					Value: r.publicHubEndpoint,
				},
			},
		},
		Inputs: wfv1.Inputs{
			Artifacts: artifacts,
		},
	}

	idx, err := GetEntrypointWorkflowIndex(rootWorkflow)
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

// AddUpdateTypeInstancesStep adds a workflow step to update TypeInstances in the Capact Local Hub.
func (r *TypeInstanceHandler) AddUpdateTypeInstancesStep(rootWorkflow *Workflow, typeInstances UpdateTypeInstances, ownerID string) error {
	artifacts := wfv1.Artifacts{}
	arguments := wfv1.Artifacts{}

	payload := []graphqllocal.UpdateTypeInstancesInput{}

	for _, ti := range typeInstances {
		artifacts = append(artifacts, wfv1.Artifact{
			Name: ti.ID,
			Path: fmt.Sprintf("/update/typeInstances/%s", ti.ID),
		})

		arguments = append(arguments, wfv1.Artifact{
			Name: ti.ID,
			From: fmt.Sprintf("{{workflow.outputs.artifacts.%s}}", ti.ArtifactName),
		})

		payload = append(payload, graphqllocal.UpdateTypeInstancesInput{
			ID:        ti.ID,
			OwnerID:   &ownerID,
			CreatedBy: &ownerID,
			TypeInstance: &graphqllocal.UpdateTypeInstanceInput{
				Attributes: []*graphqllocal.AttributeReferenceInput{},
			},
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
		Path: "/update/payload",
	})

	template := &wfv1.Template{
		Name: "upload-update-type-instances",
		Container: &apiv1.Container{
			Image:           r.hubActionsImage,
			ImagePullPolicy: apiv1.PullIfNotPresent,
			Env: []apiv1.EnvVar{
				{
					Name:  "APP_ACTION",
					Value: "UpdateAction",
				},
				{
					Name:  "APP_UPDATE_CONFIG_PAYLOAD_FILEPATH",
					Value: "/update/payload",
				},
				{
					Name:  "APP_UPDATE_CONFIG_TYPE_INSTANCES_DIR",
					Value: "/update/typeInstances",
				},
				{
					Name:  "APP_LOCAL_HUB_ENDPOINT",
					Value: r.localHubEndpoint,
				},
				{
					Name:  "APP_PUBLIC_HUB_ENDPOINT",
					Value: r.publicHubEndpoint,
				},
			},
		},
		Inputs: wfv1.Inputs{
			Artifacts: artifacts,
		},
	}

	idx, err := GetEntrypointWorkflowIndex(rootWorkflow)
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
