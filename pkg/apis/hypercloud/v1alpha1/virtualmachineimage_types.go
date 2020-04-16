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
	// VirtualMachineImageStatePvcCreating indicates pvc for VirtualMachineImage is creating
	VirtualMachineImageStatePvcCreating VirtualMachineImageState = "PvcCreating"
	// VirtualMachineImageStatePvcCreatingError indicates pvc creating is error
	VirtualMachineImageStatePvcCreatingError VirtualMachineImageState = "PvcCreatingError"
	// VirtualMachineImageStateAvailable indicates VirtualMachineImage is available
	VirtualMachineImageStateAvailable VirtualMachineImageState = "Available"
	// VirtualMachineImageStateError indicates VirtualMachineImage is error
	VirtualMachineImageStateError VirtualMachineImageState = "Error"
	// VirtualMachineImageStateImporting (TODO)
	VirtualMachineImageStateImporting VirtualMachineImageState = "Importing"
	// VirtualMachineImageStateImportingError (TODO)
	VirtualMachineImageStateImportingError VirtualMachineImageState = "ImportingError"
	// VirtualMachineImageStateSnapshotting (TODO)
	VirtualMachineImageStateSnapshotting VirtualMachineImageState = "Snapshotting"
	// VirtualMachineImageStateSnapshottingError (TODO)
	VirtualMachineImageStateSnapshottingError VirtualMachineImageState = "SnapshottingError"
)

// VirtualMachineImageStatus defines the observed state of VirtualMachineImage
type VirtualMachineImageStatus struct {
	// State is the current state of VirtualMachineImage
	State VirtualMachineImageState `json:"state"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualMachineImage is the Schema for the virtualmachineimages API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=virtualmachineimages,scope=Namespaced
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
