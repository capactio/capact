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

	// ActionRef contains data sufficient to resolve Implementation or Interface manifest.
	// Currently only Interface reference is supported.
	ActionRef ManifestReference `json:"actionRef,omitempty"`

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

	// DryRun specifies whether runner should perform only dry-run action without persisting the resource.
	// +optional
	// +kubebuilder:default=false
	DryRun *bool `json:"dryRun,omitempty"`

	// Cancel specifies whether the Action execution should be canceled.
	// +optional
	// +kubebuilder:default=false
	Cancel *bool `json:"cancel,omitempty"`
}

func isBoolSet(in *bool) bool {
	return in != nil && *in
}

func (in *ActionSpec) IsDryRun() bool {
	return isBoolSet(in.DryRun)
}

func (in *ActionSpec) IsRun() bool {
	return isBoolSet(in.Run)
}

func (in *ActionSpec) IsCanceled() bool {
	return isBoolSet(in.Cancel)
}

func (in *ActionSpec) IsAdvancedRenderingEnabled() bool {
	return in.AdvancedRendering != nil && in.AdvancedRendering.Enabled
}

func (in *Action) IsExecuted() bool {
	return in.Status.Phase == RunningActionPhase || in.Status.Phase == BeingCanceledActionPhase
}

// TODO bug, that newly created Action CR has empty phase and not the default, so we need to handle it here
func (in *Action) IsUninitialized() bool {
	return in.Status.Phase == "" || in.Status.Phase == InitialActionPhase
}

func (in *Action) IsBeingRendered() bool {
	return in.Status.Phase == BeingRenderedActionPhase
}

// IsWaitingToRun returns true if Action is fully rendered and waiting for user approval.
func (in *Action) IsWaitingToRun() bool {
	return in.Status.Phase == ReadyToRunActionPhase && !in.Spec.IsRun()
}

// IsReadyToExecute returns true if Action is fully rendered and approved by user.
func (in *Action) IsReadyToExecute() bool {
	return in.Status.Phase == ReadyToRunActionPhase && in.Spec.IsRun()
}

// IsBeingDeleted returns true if a deletion timestamp is set
func (in *Action) IsBeingDeleted() bool {
	return !in.ObjectMeta.DeletionTimestamp.IsZero()
}

// ActionInput describes Action input.
type ActionInput struct {

	// TypeInstances contains input TypeInstances passed for Action rendering. It contains both required and optional input TypeInstances.
	// +optional
	TypeInstances *[]InputTypeInstance `json:"typeInstances,omitempty"`

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

	// ApprovedIterationName specifies the name of rendering iteration, which has been approved by user.
	// Iteration approval is the user intention to continue rendering using the provided ActionInput.typeInstances in the Action input.
	// User may or may not add additional optional TypeInstances to the list and continue Action rendering.
	ApprovedIterationName string `json:"approvedIterationName"`
}

// InputTypeInstance holds input TypeInstance reference.
type InputTypeInstance struct {

	// Name refers to input TypeInstance name used in rendered Action.
	// Name is not unique as there may be multiple TypeInstances with the same name on different levels of Action workflow.
	Name string `json:"name"`

	// ID is a unique identifier for the input TypeInstance.
	ID string `json:"id"`
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

	// CanceledBy holds user data which canceled a given Action.
	// +optional
	CanceledBy *authv1.UserInfo `json:"canceledBy,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed Action.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// ActionOutput describes Action output.
type ActionOutput struct {

	// TypeInstances contains output TypeInstances data.
	// +optional
	TypeInstances *[]OutputTypeInstanceDetails `json:"typeInstances,omitempty"`
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

func (r *RenderingStatus) SetAction(action []byte) {
	r.Action = &runtime.RawExtension{Raw: action}
}

func (r *RenderingStatus) SetInputParameters(params []byte) {
	if r.Input == nil {
		r.Input = &ResolvedActionInput{}
	}
	r.Input.SetParameters(params)
}

// ResolvedActionInput contains resolved details of Action input.
type ResolvedActionInput struct {
	// TypeInstances contains input TypeInstances passed for Action rendering.
	// It contains both required and optional input TypeInstances.
	// +optional
	TypeInstances *[]InputTypeInstanceDetails `json:"typeInstances,omitempty"`

	// Parameters holds value of the User input parameters.
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Parameters *runtime.RawExtension `json:"parameters,omitempty"`
}

func (r *ResolvedActionInput) SetParameters(params []byte) {
	r.Parameters = &runtime.RawExtension{Raw: params}
}

// InputTypeInstanceDetails describes input TypeInstance.
type InputTypeInstanceDetails struct {

	// TODO: After first implementation of rendering workflow, make Input TypeInstance unique.
	// Possible options:
	// - name prefix is added manually by User during advanced rendering
	// - introduce additional field `prefix` or `location`, `source`, etc. with path to the nested step
	// - similarly to Argo, add special steps with children data

	CommonTypeInstanceDetails `json:",inline"`

	// Optional highlights that the input TypeInstance is optional.
	// +kubebuilder:default=false
	Optional bool `json:"optional,omitempty"`
}

// OutputTypeInstanceDetails describes the output TypeInstance.
type OutputTypeInstanceDetails struct {
	CommonTypeInstanceDetails `json:",inline"`
}

// CommonTypeInstanceDetails contains common details of TypeInstance, regardless if it is input or output one.
type CommonTypeInstanceDetails struct {

	// Name refers to TypeInstance name.
	Name string `json:"name"`

	// ID is a unique identifier of the TypeInstance.
	ID string `json:"id"`

	// TypeRef contains data needed to resolve Type manifest.
	TypeRef *ManifestReference `json:"typeReference"`
}

// InputTypeInstanceToProvide describes optional input TypeInstance for advanced rendering mode iteration.
type InputTypeInstanceToProvide struct {

	// Name refers to TypeInstance name.
	Name string `json:"name"`

	// TypeRef contains data needed to resolve Type manifest.
	TypeRef *ManifestReference `json:"typeReference"`
}

// ManifestReference contains data needed to resolve a manifest.
type ManifestReference struct {

	// Path is full path for the manifest.
	Path NodePath `json:"path"`

	// Revision is a semantic version of the manifest. If not provided, the latest revision is used.
	// +optional
	Revision *string `json:"revision,omitempty"`
}

// AdvancedRenderingStatus describes status related to advanced rendering mode.
type AdvancedRenderingStatus struct {

	// RenderingIteration describes status related to current rendering iteration.
	// +optional
	RenderingIteration *RenderingIterationStatus `json:"renderingIteration,omitempty"`
}

// RenderingIterationStatus holds status for current rendering iteration in advanced rendering mode.
type RenderingIterationStatus struct {

	// CurrentIterationName contains name of current iteration in advanced rendering.
	CurrentIterationName string `json:"currentIterationName"`

	// InputTypeInstancesToProvide describes which input TypeInstances might be provided in a given rendering iteration.
	// +optional
	InputTypeInstancesToProvide *[]InputTypeInstanceToProvide `json:"inputTypeInstancesToProvide,omitempty"`
}

// RunnerStatus holds data related to built-in Runner that runs the Action.
type RunnerStatus struct {

	// TODO: Once we will support nested runners statues, add Interface property which is a full path of Runner Interface manifest .

	// Status contains reference to resource with arbitrary Runner status data.
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Status *runtime.RawExtension `json:"status,omitempty"`
}

// NodePath defines full path for a given manifest, e.g. Implementation or Interface.
// +kubebuilder:validation:MinLength=3
type NodePath string

// ActionPhase describes in which state is the Action to execute.
// +kubebuilder:validation:Enum=Initial;BeingRendered;AdvancedModeRenderingIteration;ReadyToRun;Running;BeingCanceled;Canceled;Succeeded;Failed
type ActionPhase string

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
