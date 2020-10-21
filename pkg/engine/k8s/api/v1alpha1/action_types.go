// +kubebuilder:validation:Required
package v1alpha1

import (
	"encoding/json"
	"k8s.io/api/authentication/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: kubebuilder printcolumn
// TODO: add validation
// TODO: To consider status conditions?
// TODO: Update example
// TODO: Update sample engine code

// NOTE: json tags are required. Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "make gen-k8s-resources" to regenerate code after modifying this file.

// ActionSpec contains configuration properties for a given Action to execute.
type ActionSpec struct {

	// Path contains full path for Implementation or Interface manifest.
	Path NodePath `json:"path,omitempty"`

	// InputRef contains reference to resource with Action input.
	// +optional
	InputRef *ActionIORef `json:"inputRef,omitempty"`

	// AdvancedRendering holds are properties related to Action advanced rendering mode.
	// +optional
	AdvancedRendering AdvancedRenderingSpec `json:"advancedRendering,omitempty"`

	// RenderedActionOverride contains optional rendered Action that overrides the one rendered by Engine.
	// +optional
	RenderedActionOverride *json.RawMessage `json:"renderedActionOverride,omitempty"`

	// Run specifies whether the Action is approved to be executed.
	// Engine won't execute fully rendered Action until the field is set to `true`.
	// If the Action is not fully rendered, and this field is set to `true`, Engine executes a given Action instantly after it is resolved.
	// +optional
	Run bool `json:"run,omitempty"`

	// Cancel specifies whether the Action execution should be cancelled.
	// +optional
	Cancel bool `json:"cancel,omitempty"`
}

// AdvancedRenderingSpec holds are properties related to Action advanced rendering mode.
type AdvancedRenderingSpec struct {

	// Enabled specifies if the advanced rendering mode is enabled.
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// RenderingIteration holds properties for rendering iteration in advanced rendering mode.
	// +optional
	RenderingIteration RenderingIterationSpec `json:"renderingIteration,omitempty"`
}

// RenderingIterationSpec holds properties for rendering iteration in advanced rendering mode.
type RenderingIterationSpec struct {

	// InputArtifacts contains Input Artifacts passed for current rendering iteration.
	// +optional
	InputArtifacts *[]InputArtifact `json:"inputArtifacts,omitempty"`
}

// InputArtifact holds input artifact reference, which is in fact TypeInstance.
type InputArtifact struct {

	// Alias refers to input artifact name used in rendered Action.
	Alias string `json:"alias,omitempty"`

	// TypePath is full path for the Type manifest related to a given artifact (TypeInstance).
	TypePath NodePath `json:"typePath,omitempty"`

	// TypeInstanceID is a unique identifier for the TypeInstance used as input artifact.
	TypeInstanceID string `json:"typeInstanceID,omitempty"`
}

// ActionIORef holds references to resources where Action input or output is stored.
type ActionIORef struct {

	// SecretRef stores reference to Secret in the same namespace the Action CR is created.
	// +optional
	SecretRef *v1.LocalObjectReference `json:"secretRef,omitempty"`
}

// ActionStatus defines the observed state of Action.
type ActionStatus struct {

	// ActionPhase describes in which state is the Action to execute.
	Phase ActionPhase `json:"phase,omitempty"`

	// Message provides a readable description of the Action state.
	// +optional
	Message *string `json:"message,omitempty"`

	// BuiltinRunner holds data related to built-in Runner that runs the Action.
	BuiltinRunner BuiltinRunnerStatus `json:"builtInRunner,omitempty"`

	// OutputRef contains reference to resource with Action output.
	// +optional
	OutputRef *ActionIORef `json:"outputRef,omitempty"`

	// RenderedAction contains partially or fully rendered Action to be executed.
	// +optional
	RenderedAction *json.RawMessage `json:"renderedAction,omitempty"`

	// AdvancedRendering describes status related to advanced rendering mode.
	AdvancedRendering *AdvancedRenderingStatus `json:"advancedRendering,omitempty"`

	// CreatedBy holds user data which created a given Action.
	// +optional
	CreatedBy *v1beta1.UserInfo `json:"createdBy,omitempty"`

	// RunBy holds user data which run a given Action.
	// +optional
	RunBy *v1beta1.UserInfo `json:"runBy,omitempty"`

	// CancelledBy holds user data which cancelled a given Action.
	// +optional
	CancelledBy *v1beta1.UserInfo `json:"cancelledBy,omitempty"`
}

// AdvancedRenderingStatus describes status related to advanced rendering mode.
type AdvancedRenderingStatus struct {

	// RenderingIteration describes status related to current rendering iteration.
	// +optional
	RenderingIteration *RenderingIterationStatus `json:"renderingIteration,omitempty"`
}

// RenderingIterationStatus holds status for current rendering iteration in advanced rendering mode.
type RenderingIterationStatus struct {

	// InputArtifactsToProvide describes which input artifacts might be provided in a given rendering iteration.
	// +optional
	InputArtifactsToProvide *[]InputArtifact `json:"inputArtifactsToProvide,omitempty"`
}

// InputArtifactsToProvide describes input artifact that may be provided in a given rendering iteration.
type InputArtifactToProvide struct {

	// Alias refers to input artifact name used in rendered Action.
	Alias string `json:"alias,omitempty"`

	// TypePath is full path for the Type manifest related to a given artifact (TypeInstance).
	TypePath NodePath `json:"typePath,omitempty"`
}

// BuiltinRunnerStatus holds data related to built-in Runner that runs the Action.
type BuiltinRunnerStatus struct {

	// Interface is a full path of built-in Runner Interface manifest.
	Interface NodePath `json:"interface,omitempty"`

	// StatusRef contains reference to resource with built-in Runner status data.
	// +optional
	Status *json.RawMessage `json:"status,omitempty"`
}

// NodePath defines full path for a given manifest, e.g. Implementation or Interface.
// +kubebuilder:validation:MinLength=3
type NodePath string

// ActionPhase describes in which state is the Action to execute.
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

// Action describes user intention to resolve & execute a given Interface or Implementation.
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
