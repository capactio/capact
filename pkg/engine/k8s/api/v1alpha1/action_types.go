package v1alpha1

import (
	"encoding/json"
	"k8s.io/api/authentication/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: kubebuilder printcolumn
// TODO: add comments to every field
// TODO: add validation
// TODO: To consider status conditions?

// NOTE: json tags are required. Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "make gen-k8s-resources" to regenerate code after modifying this file.

// ActionSpec describes user intention to resolve & execute a given Interface or Implementation.
type ActionSpec struct {

	// Full path for the Implementation or Interface manifest
	Path ActionPath `json:"path,omitempty"`

	// InputRef contains reference to resource with Action input.
	InputRef  *ActionIORef `json:"inputRef,omitempty"`

	AdvancedRendering AdvancedRenderingSpec `json:"advancedRendering,omitempty"`

	RenderedActionOverride *json.RawMessage `json:"renderedActionOverride,omitempty"`

	Run bool `json:"run,omitempty"`
}

type AdvancedRenderingSpec struct {

	Enabled bool `json:"enabled,omitempty"`

	RenderingIteration RenderingIterationSpec `json:"renderingIteration,omitempty"`
}

type RenderingIterationSpec struct {

	InputArtifacts *[]InputArtifact `json:"inputArtifacts,omitempty"`
}

type InputArtifact struct {

	Alias string `json:"alias,omitempty"`

	TypePath NodePath `json:"typePath,omitempty"`

	TypeInstanceID *string `json:"typeInstanceID,omitempty"`

}

// ActionPath defines full path for the Implementation or Interface to run.
type ActionPath string

type ActionIORef struct {
	SecretRef *v1.LocalObjectReference `json:"secretRef,omitempty"`
}

// ActionStatus defines the observed state of Action.
type ActionStatus struct {

	Phase ActionPhase `json:"phase,omitempty"`

	Message *string `json:"message,omitempty"`

	BuiltinRunner BuiltinRunnerStatus `json:"builtInRunner,omitempty"`

	// OutputRef contains reference to resource with Action output.
	OutputRef *ActionIORef `json:"outputRef,omitempty"`

	RenderedAction *json.RawMessage `json:"renderedAction,omitempty"`

	AdvancedRendering *AdvancedRenderingStatus `json:"advancedRendering,omitempty"`

	CreatedBy   *v1beta1.UserInfo `json:"createdBy,omitempty"`
	RunBy       *v1beta1.UserInfo `json:"runBy,omitempty"`
	CancelledBy *v1beta1.UserInfo `json:"cancelledBy,omitempty"`
}

type AdvancedRenderingStatus struct {
	RenderingIteration *RenderingIterationStatus `json:"renderingIteration,omitempty"`
}

type RenderingIterationStatus struct {
	InputArtifactsToProvide *[]InputArtifact `json:"inputArtifactsToProvide,omitempty"`
}

type InputArtifactToProvide struct {
	Alias string `json:"alias,omitempty"`
	TypePath NodePath `json:"typePath,omitempty"`
}

// BuiltinRunnerStatus holds data related to built-in Runner that runs the Action.
type BuiltinRunnerStatus struct {

	// Interface is a full path of built-in Runner Interface manifest.
	Interface NodePath `json:"interface,omitempty"`

	// StatusRef contains reference to resource with built-in Runner status data.
	Status *json.RawMessage `json:"status,omitempty"`
}

type NodePath string

type ActionPhase string

const (
	InitialActionPhase                        ActionPhase = "Initial"
	BeingRenderedActionPhase                  ActionPhase = "BeingRendered"
	AdvancedModeRenderingIterationActionPhase ActionPhase = "AdvancedModeRenderingIteration"
	ReadyToRunActionPhase                     ActionPhase = "ReadyToRun"
	RunningActionPhase                        ActionPhase = "Running"
	BeingCancelledActionPhase                 ActionPhase = "BeingCancelled"
	CancelledActionPhase                      ActionPhase = "Cancelled"
	SucceededActionPhase                      ActionPhase = "Succeeded"
	FailedActionPhase                         ActionPhase = "Failed"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Action is the Schema for the actions API
type Action struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ActionSpec   `json:"spec,omitempty"`
	Status ActionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ActionList contains a list of Action
type ActionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Action `json:"items"`
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&Action{}, &ActionList{})
}
