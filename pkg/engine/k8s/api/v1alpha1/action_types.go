// +kubebuilder:validation:Required
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/api/authentication/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required. Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "make gen-k8s-resources" to regenerate code after modifying this file.

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=ac
// +kubebuilder:printcolumn:name="Path",type="string",JSONPath=".spec.path",description="Interface/Implementation path of the Action"
// +kubebuilder:printcolumn:name="Run",type="boolean",JSONPath=".spec.run",description="If the Action is approved to run"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase",description="Status of the Action"
// +kubebuilder:printcolumn:name="Age",type="date",format="date-time",JSONPath=".metadata.creationTimestamp",description="When the Action was created"

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

// ActionSpec contains configuration properties for a given Action to execute.
type ActionSpec struct {

	// Path contains full path for Implementation or Interface manifest.
	Path NodePath `json:"path"`

	// InputRef contains reference to resource with Action input.
	// +optional
	InputRef *ActionIORef `json:"inputRef,omitempty"`

	// AdvancedRendering holds properties related to Action advanced rendering mode.
	// +optional
	AdvancedRendering *AdvancedRendering `json:"advancedRendering,omitempty"`

	// RenderedActionOverride contains optional rendered Action that overrides the one rendered by Engine.
	// +optional
	RenderedActionOverride *runtime.RawExtension `json:"renderedActionOverride,omitempty"`

	// Run specifies whether the Action is approved to be executed.
	// Engine won't execute fully rendered Action until the field is set to `true`.
	// If the Action is not fully rendered, and this field is set to `true`, Engine executes a given Action instantly after it is resolved.
	// +optional
	// +kubebuilder:default=false
	Run *bool `json:"run,omitempty"`

	// Cancel specifies whether the Action execution should be cancelled.
	// +optional
	// +kubebuilder:default=false
	Cancel *bool `json:"cancel,omitempty"`
}

// AdvancedRendering holds are properties related to Action advanced rendering mode.
type AdvancedRendering struct {

	// Enabled specifies if the advanced rendering mode is enabled.
	// +kubebuilder:default=false
	Enabled bool `json:"enabled"`

	// RenderingIteration holds properties for rendering iteration in advanced rendering mode.
	// +optional
	RenderingIteration *RenderingIteration `json:"renderingIteration,omitempty"`
}

// RenderingIteration holds properties for rendering iteration in advanced rendering mode.
type RenderingIteration struct {

	// InputArtifacts contains Input Artifacts passed for current rendering iteration.
	// +optional
	InputArtifacts *[]InputArtifact `json:"inputArtifacts,omitempty"`

	// Continue specifies the user intention to continue rendering using the provided InputArtifacts.
	// As the input artifacts are optional, user may continue rendering with empty list of InputArtifacts.
	// +kubebuilder:default=false
	Continue bool `json:"continue"`
}

// InputArtifact holds input artifact reference, which is in fact TypeInstance.
type InputArtifact struct {

	// Alias refers to input artifact name used in rendered Action.
	Alias string `json:"alias"`

	// TypeInstanceID is a unique identifier for the TypeInstance used as input artifact.
	TypeInstanceID string `json:"typeInstanceID"`
}

// ActionIORef holds references to resources where Action input or output is stored.
type ActionIORef struct {

	// SecretRef stores reference to Secret in the same namespace the Action CR is created.
	SecretRef v1.LocalObjectReference `json:"secretRef"`
}

// ActionStatus defines the observed state of Action.
type ActionStatus struct {

	// TODO: To investigate why the status phase is not initially filled with the default value; OpenAPI schema is correctly rendered

	// ActionPhase describes in which state is the Action to execute.
	// +kubebuilder:default=Initial
	Phase ActionPhase `json:"phase"`

	// Message provides a readable description of the Action phase.
	// +optional
	Message *string `json:"message,omitempty"`

	// Runner holds data related to Runner that runs the Action.
	// +optional
	Runner *RunnerStatus `json:"runner,omitempty"`

	// OutputRef contains reference to resource with Action output.
	// +optional
	OutputRef *ActionIORef `json:"outputRef,omitempty"`

	// RenderedAction contains partially or fully rendered Action to be executed.
	// +optional
	RenderedAction *runtime.RawExtension `json:"renderedAction,omitempty"`

	// AdvancedRendering describes status related to advanced rendering mode.
	// +optional
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
	// +kubebuilder:validation:MinLength=1
	Alias string `json:"alias"`

	// TypePath is full path for the Type manifest related to a given artifact (TypeInstance).
	TypePath NodePath `json:"typePath"`
}

// RunnerStatus holds data related to built-in Runner that runs the Action.
type RunnerStatus struct {

	// Interface is a full path of Runner Interface manifest.
	Interface NodePath `json:"interface"`

	// StatusRef contains reference to resource with arbitrary Runner status data.
	// +optional
	Status *runtime.RawExtension `json:"status,omitempty"`
}

// NodePath defines full path for a given manifest, e.g. Implementation or Interface.
// +kubebuilder:validation:MinLength=3
type NodePath string

// ActionPhase describes in which state is the Action to execute.
// +kubebuilder:validation:Enum=Initial;BeingRendered;AdvancedModeRenderingIteration;ReadyToRun;Running;BeingCancelled;Cancelled;Succeeded;Failed
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
