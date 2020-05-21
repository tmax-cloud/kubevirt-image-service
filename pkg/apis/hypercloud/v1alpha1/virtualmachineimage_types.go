package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VirtualMachineImageSource provides parameters to create a VirtualMachineImage from an HTTP source
type VirtualMachineImageSource struct {
	HTTP string `json:"http"`
}

// VirtualMachineImageSpec defines the desired state of VirtualMachineImage
type VirtualMachineImageSpec struct {
	Source            VirtualMachineImageSource        `json:"source"`
	PVC               corev1.PersistentVolumeClaimSpec `json:"pvc"`
	SnapshotClassName string                           `json:"snapshotClassName"`
}

// VirtualMachineImageState is the current state of VirtualMachineImage
type VirtualMachineImageState string

const (
	// VirtualMachineImageStateCreating indicates VirtualMachineImage is creating
	VirtualMachineImageStateCreating VirtualMachineImageState = "Creating"
	// VirtualMachineImageStateAvailable indicates VirtualMachineImage is available
	VirtualMachineImageStateAvailable VirtualMachineImageState = "Available"
	// VirtualMachineImageStateError indicates VirtualMachineImage is error
	VirtualMachineImageStateError VirtualMachineImageState = "Error"
)

// VirtualMachineImageConditionType defines the condition of VirtualMachineImage
type VirtualMachineImageConditionType string

// VirtualMachineImageStatus defines the observed state of VirtualMachineImage
type VirtualMachineImageStatus struct {
	// State is the current state of VirtualMachineImage
	State VirtualMachineImageState `json:"state"`
	// +optional
	ReadyToUse *bool `json:"readyToUse,omitempty"`
	// Human-readable message indicating details about last transition.
	// +optional
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualMachineImage is the Schema for the virtualmachineimages API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=virtualmachineimages,scope=Namespaced,shortName=vmim
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="Current state of VirtualMachineImage"
type VirtualMachineImage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtualMachineImageSpec   `json:"spec,omitempty"`
	Status VirtualMachineImageStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualMachineImageList contains a list of VirtualMachineImage
type VirtualMachineImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualMachineImage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtualMachineImage{}, &VirtualMachineImageList{})
}
