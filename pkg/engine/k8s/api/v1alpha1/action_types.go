// +kubebuilder:validation:Required
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"

	authv1 "k8s.io/api/authentication/v1"
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

	// Input describes Action input.
	// +optional
	Input *ActionInput `json:"input,omitempty"`

	// AdvancedRendering holds properties related to Action advanced rendering mode.
	// +optional
	AdvancedRendering *AdvancedRendering `json:"advancedRendering,omitempty"`

	// RenderedActionOverride contains optional rendered Action that overrides the one rendered by Engine.
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
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

func (in *ActionSpec) IsRun() bool {
	return in.Run != nil && *in.Run
}

func (in *ActionSpec) IsCancelled() bool {
	return in.Cancel != nil && *in.Cancel
}

// ActionInput describes Action input.
type ActionInput struct {

	// Artifacts contains input Artifacts passed for Action rendering. It contains both required and optional input Artifacts.
	// +optional
	Artifacts *[]InputArtifact `json:"artifacts,omitempty"`

	// Parameters holds details about Action input parameters.
	// +optional
	Parameters *InputParameters `json:"parameters,omitempty"`
}

// InputParameters holds details about Action input parameters.
type InputParameters struct {

	// SecretRef stores reference to Secret in the same namespace the Action CR is created.
	SecretRef v1.LocalObjectReference `json:"secretRef"`
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

	// Continue specifies the user intention to continue rendering using the provided InputArtifacts in the Action input.
	// User may or may not add additional optional InputArtifacts to the list and continue Action rendering.
	// +kubebuilder:default=false
	Continue bool `json:"continue"`
}

// InputArtifact holds input artifact reference, which is in fact TypeInstance.
type InputArtifact struct {

	// Name refers to input artifact name used in rendered Action.
	// Name is not unique as there may be multiple artifacts with the same name on different levels of Action workflow.
	Name string `json:"name"`

	// TypeInstanceID is a unique identifier for the TypeInstance used as input artifact.
	TypeInstanceID string `json:"typeInstanceID"`
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

	// Output describes Action output.
	// +optional
	Output *ActionOutput `json:"output,omitempty"`

	// Rendering describes rendering status.
	// +optional
	Rendering *RenderingStatus `json:"rendering,omitempty"`

	// CreatedBy holds user data which created a given Action.
	// +optional
	CreatedBy *authv1.UserInfo `json:"createdBy,omitempty"`

	// RunBy holds user data which run a given Action.
	// +optional
	RunBy *authv1.UserInfo `json:"runBy,omitempty"`

	// CancelledBy holds user data which cancelled a given Action.
	// +optional
	CancelledBy *authv1.UserInfo `json:"cancelledBy,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed Action.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// ActionOutput describes Action output.
type ActionOutput struct {

	// Artifacts contains output Artifacts information.
	// +optional
	Artifacts *[]OutputArtifactDetails `json:"artifacts,omitempty"`
}

// RenderingStatus describes rendering status.
type RenderingStatus struct {

	// Action contains partially or fully rendered Action to be executed.
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Action *runtime.RawExtension `json:"action,omitempty"`

	// Input contains resolved details of Action input.
	// +optional
	Input *ResolvedActionInput `json:"input,omitempty"`

	// AdvancedRendering describes status related to advanced rendering mode.
	// +optional
	AdvancedRendering *AdvancedRenderingStatus `json:"advancedRendering,omitempty"`
}

// ResolvedActionInput contains resolved details of Action input.
type ResolvedActionInput struct {
	// Artifacts contains input Artifacts passed for Action rendering. It contains both required and optional input Artifacts.
	// +optional
	Artifacts *[]InputArtifactDetails `json:"artifacts,omitempty"`

	// Parameters holds value of the User input parameters.
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Parameters *runtime.RawExtension `json:"parameters,omitempty"`
}

type InputArtifactDetails struct {

	// TODO: After first implementation of rendering workflow, make Input Artifact unique.
	// Possible options:
	// - name prefix is added manually by User during advanced rendering
	// - introduce additional field `prefix` or `location`, `source`, etc. with path to the nested step
	// - similarly to Argo, add special steps with children data

	CommonArtifactDetails `json:",inline"`

	// Optional highlights that the input artifact is optional.
	// +kubebuilder:default=false
	Optional bool `json:"optional,omitempty"`
}

type OutputArtifactDetails struct {
	CommonArtifactDetails `json:",inline"`
}

type CommonArtifactDetails struct {

	// Name refers to artifact name.
	Name string `json:"name"`

	// TypeInstanceID is a unique identifier for the TypeInstance used as artifact.
	TypeInstanceID string `json:"typeInstanceID"`

	// TypePath is full path for the Type manifest related to a given artifact (TypeInstance).
	TypePath NodePath `json:"typePath"`
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
	InputArtifactsToProvide *[]InputArtifactDetails `json:"inputArtifactsToProvide,omitempty"`
}

// RunnerStatus holds data related to built-in Runner that runs the Action.
type RunnerStatus struct {

	// Interface is a full path of Runner Interface manifest.
	Interface NodePath `json:"interface"`

	// StatusRef contains reference to resource with arbitrary Runner status data.
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
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
