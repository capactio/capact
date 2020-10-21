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
// TODO: conditions?

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "make gen-k8s-resources" to regenerate code after modifying this file

// ActionSpec describes user intention to resolve & execute a given Interface or Implementation.
type ActionSpec struct {

	// Full path for the Implementation or Interface
	Action ActionPath `json:"action,omitempty"`

	// InputRef contains reference to resource with Action input.
	InputRef  *ActionIORef `json:"inputRef,omitempty"`

	AdvancedRendering AdvancedRenderingSpec `json:"advancedRendering,omitempty"`

	RenderedActionOverride *json.RawMessage

	Run bool
}

type AdvancedRenderingSpec struct {
	Enabled bool `json:"advancedRenderingEnabled,omitempty"`
	RenderingIteration RenderingIterationSpec
}

type RenderingIterationSpec struct {
	InputArtifacts *[]InputArtifact `json:"artifactInput,omitempty"`
}

type InputArtifact struct {
	Alias string
	TypePath NodePath
	TypeInstanceID *string
}

// ActionPath defines full path for the Implementation or Interface to run.
type ActionPath string

type ActionIORef struct {
	SecretRef *v1.LocalObjectReference `json:"secretRef,omitempty"`
}

// ActionStatus defines the observed state of Action.
type ActionStatus struct {
	Condition ActionCondition `json:"condition"`

	Message *string `json:"message,omitempty"`

	BuiltinRunner BuiltinRunnerStatus `json:"builtInRunner,omitempty"`

	// OutputRef contains reference to resource with Action output.
	OutputRef *ActionIORef `json:"outputRef,omitempty"`

	RenderedAction *json.RawMessage

	AdvancedRendering *AdvancedRenderingStatus `json:"renderingAdvancedMode,omitempty"`

	CreatedBy   *v1beta1.UserInfo `json:"createdBy,omitempty"`
	RunBy       *v1beta1.UserInfo `json:"runBy,omitempty"`
	CancelledBy *v1beta1.UserInfo `json:"cancelledBy,omitempty"`
}

type AdvancedRenderingStatus struct {
	RenderingIteration *RenderingIterationStatus
}

type RenderingIterationStatus struct {
	InputArtifactsToProvide *[]InputArtifact
}

type InputArtifactToProvide struct {
	Alias string
	TypePath NodePath
}

type BuiltinRunnerStatus struct {

	Interface NodePath `json:"interface,omitempty"`

	// StatusRef contains reference to resource with built-in Runner status data.
	Status *json.RawMessage `json:"status,omitempty"`
}

type NodePath string

type ActionCondition string

const (
	InitialActionCondition                        ActionCondition = "Initial"
	BeingRenderedActionCondition                  ActionCondition = "BeingRendered"
	AdvancedModeRenderingIterationActionCondition ActionCondition = "AdvancedModeRenderingIteration"
	ReadyToRunActionCondition                     ActionCondition = "ReadyToRun"
	RunningActionCondition                        ActionCondition = "Running"
	BeingCancelledActionCondition                 ActionCondition = "BeingCancelled"
	CancelledActionCondition                      ActionCondition = "Cancelled"
	SucceededActionCondition                      ActionCondition = "Succeeded"
	FailedActionCondition                         ActionCondition = "Failed"
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
