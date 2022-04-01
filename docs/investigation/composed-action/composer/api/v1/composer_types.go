/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Composer is the Schema for the composers API
type Composer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComposerSpec   `json:"spec,omitempty"`
	Status ComposerStatus `json:"status,omitempty"`
}

func (in *Composer) IsUninitialized() bool {
	return in.Status.Phase == "" || in.Status.Phase == InitialComposerPhase
}

func (in *Composer) ScheduleNextIdx() int {
	nextItemIdx := func(idx int) int {
		idx += 1
		if idx >= len(in.Status.Results) {
			return -1
		}
		return idx
	}

	var lastSucceededIDDx int

	for idx, item := range in.Status.Results {
		if item.Phase == FailedComposerPhase {
			return -1 // cancel all next iteration
		}

		if item.Phase == RunningComposerPhase {
			return -1 // still running
		}

		if item.Phase == SucceededComposerPhase {
			lastSucceededIDDx = idx
		}
	}

	return nextItemIdx(lastSucceededIDDx)
}

func (in *Composer) IsRunning() bool {
	for _, item := range in.Status.Results {
		if item.Phase == RunningComposerPhase {
			return true
		}
	}
	return false
}

//+kubebuilder:object:root=true

// ComposerList contains a list of Composer
type ComposerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Composer `json:"items"`
}

// ComposerSpec defines the desired state of Composer
type ComposerSpec struct {
	Steps map[string]Step `json:"step,omitempty"`
	// +optional
	Input *ActionInput `json:"input,omitempty"`
}

type Step struct {
	Interface InterfaceRef         `json:"interface"`
	Input     map[string]StepInput `json:"input"`
}

type StepInput struct {
	From *string `json:"from,omitempty"`
	Raw  *string `json:"raw,omitempty"`
}

type InterfaceRef struct {
	Path string `json:"path"`
}

// ComposerStatus defines the observed state of Composer
type ComposerStatus struct {
	StartTime      *metav1.Time `json:"startTime,inline,omitempty"`
	CompletionTime *metav1.Time `json:"completionTime,inline,omitempty"`
	// +kubebuilder:default=Initial
	Phase   ComposerPhase    `json:"phase"`
	Results []ComposerResult `json:"results,omitempty"`
}

// +kubebuilder:validation:Enum=Initial;Running;Succeeded;Failed
type ComposerPhase string

const (
	InitialComposerPhase   ComposerPhase = "Initial"
	RunningComposerPhase   ComposerPhase = "Running"
	SucceededComposerPhase ComposerPhase = "Succeeded"
	FailedComposerPhase    ComposerPhase = "Failed"
)

type ComposerResult struct {
	Name           string        `json:"name"`
	Phase          ComposerPhase `json:"phase"`
	StartTime      *metav1.Time  `json:"startTime,inline,omitempty"`
	CompletionTime *metav1.Time  `json:"completionTime,inline,omitempty"`

	// Action related (copied)

	ActionPhase ActionPhase `json:"actionPhase"`
	// +optional
	Message *string `json:"message,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Composer{}, &ComposerList{})
}

// Copied from Action

// ActionPhase describes in which state is the Action to execute.
// +kubebuilder:validation:Enum=Initial;BeingRendered;AdvancedModeRenderingIteration;ReadyToRun;Running;BeingCanceled;Canceled;Succeeded;Failed
type ActionPhase string

// List of possible Action phases.
const (
	InitialActionPhase                        ActionPhase = "Initial"
	BeingRenderedActionPhase                  ActionPhase = "BeingRendered"
	AdvancedModeRenderingIterationActionPhase ActionPhase = "AdvancedModeRenderingIteration"
	ReadyToRunActionPhase                     ActionPhase = "ReadyToRun"
	RunningActionPhase                        ActionPhase = "Running"
	BeingCanceledActionPhase                  ActionPhase = "BeingCanceled"
	CanceledActionPhase                       ActionPhase = "Canceled"
	SucceededActionPhase                      ActionPhase = "Succeeded"
	FailedActionPhase                         ActionPhase = "Failed"
)

// ActionInput describes Action input.
type ActionInput struct {

	// TypeInstances contains required input TypeInstances passed for Action rendering.
	// +optional
	TypeInstances *[]InputTypeInstance `json:"typeInstances,omitempty"`

	// Parameters holds details about Action input parameters.
	// +optional
	Parameters *InputParameters `json:"parameters,omitempty"`

	// Describes the one-time User policy.
	// +optional
	ActionPolicy *ActionPolicy `json:"policy,omitempty"`
}

// InputTypeInstance holds input TypeInstance reference.
type InputTypeInstance struct {

	// Name refers to input TypeInstance name used in rendered Action.
	// Name is not unique as there may be multiple TypeInstances with the same name on different levels of Action workflow.
	Name string `json:"name"`

	// ID is a unique identifier for the input TypeInstance.
	ID string `json:"id"`
}

// InputParameters holds details about Action input parameters.
type InputParameters struct {

	// SecretRef stores reference to Secret in the same namespace the Action CR is created.
	//
	// Required field:
	// - Secret.Data["parameters.json"] - input parameters data in JSON format
	//
	// Restricted field:
	// - Secret.Data["args.yaml"] - used by Engine, stores runner rendered arguments
	// - Secret.Data["context.yaml"] - used by Engine, stores runner context
	// - Secret.Data["status"] - stores the runner status
	// - Secret.Data["action-policy.json"] - stores the one-time Action policy in JSON format
	//
	// TODO: this should be changed to an object which contains both the Secret name and key
	// name under which the input is stored.
	SecretRef v1.LocalObjectReference `json:"secretRef"`
}

// ActionPolicy describes Action Policy reference.
type ActionPolicy struct {

	// SecretRef stores reference to Secret in the same namespace the Action CR is created.
	//
	// Required field:
	// - Secret.Data["action-policy.json"] - stores the one-time Action policy in JSON format
	//
	// Restricted field:
	// - Secret.Data["args.yaml"] - used by Engine, stores runner rendered arguments
	// - Secret.Data["context.yaml"] - used by Engine, stores runner context
	// - Secret.Data["status"] - stores the runner status
	// - Secret.Data["parameters.json"] - input parameters data in JSON format
	//
	SecretRef v1.LocalObjectReference `json:"secretRef"`
}
