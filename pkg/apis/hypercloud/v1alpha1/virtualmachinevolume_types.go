package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VirtualMachineImageName identifies which VirtualMachineImage source to create a VirtualMachineVolume from
type VirtualMachineImageName struct {
	Name string `json:"name"`
}

// ResourceName is the name identifying various resources in a ResourceList.
type ResourceName string

// VirtualMachineVolumeSpec defines the desired state of VirtualMachineVolume
type VirtualMachineVolumeSpec struct {
	// VirtualMachineImage defines name of the VirtualMachineImage
	VirtualMachineImage VirtualMachineImageName `json:"virtualMachineImage"`
	// Capacity defines size of the VirtualMachineVolume
	Capacity corev1.ResourceList `json:"capacity,omitempty" protobuf:"bytes,1,rep,name=capacity,casttype=ResourceList,castkey=ResourceName"`
}

// VirtualMachineVolumeState is the current state of VirtualMachineVolume
type VirtualMachineVolumeState string

const (
	// VirtualMachineVolumeStateCreating indicates VirtualMachineVolume is creating pvc
	VirtualMachineVolumeStateCreating VirtualMachineVolumeState = "Creating"
	// VirtualMachineVolumeStateAvailable indicates VirtualMachineVolume is ready to use
	VirtualMachineVolumeStateAvailable VirtualMachineVolumeState = "Available"
	// VirtualMachineVolumeStateError indicates VirtualMachineVolume is not able to use
	VirtualMachineVolumeStateError VirtualMachineVolumeState = "Error"
)

// VirtualMachineVolumeStatus defines the observed state of VirtualMachineVolume
type VirtualMachineVolumeStatus struct {
	// State is the current state of VirtualMachineVolume
	State VirtualMachineVolumeState `json:"state"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualMachineVolume is the Schema for the virtualmachinevolumes API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=virtualmachinevolumes,scope=Namespaced
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="Current state of VirtualMachineVolume"
type VirtualMachineVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtualMachineVolumeSpec   `json:"spec,omitempty"`
	Status VirtualMachineVolumeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualMachineVolumeList contains a list of VirtualMachineVolume
type VirtualMachineVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualMachineVolume `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtualMachineVolume{}, &VirtualMachineVolumeList{})
}
