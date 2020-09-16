package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VirtualMachineVolumeExportSpec defines the desired state of VirtualMachineVolumeExport
type VirtualMachineVolumeExportSpec struct {
	VirtualMachineVolume VirtualMachineVolumeSource            `json:"virtualMachineVolume"`
	Destination          VirtualMachineVolumeExportDestination `json:"destination"`
}

// VirtualMachineVolumeExportDestination defines the destination to export volume
type VirtualMachineVolumeExportDestination struct {
	Local *VirtualMachineVolumeExportDestinationLocal `json:"local,omitempty"`
}

// VirtualMachineVolumeSource indicates the VirtualMachineVolume to be exported
type VirtualMachineVolumeSource struct {
	Name string `json:"name"`
}

// VirtualMachineVolumeExportDestinationLocal defines the Local destination to export volume
type VirtualMachineVolumeExportDestinationLocal struct{}

// VirtualMachineVolumeExportStatus defines the observed state of VirtualMachineVolumeExport
type VirtualMachineVolumeExportStatus struct {
	// State is the current state of VirtualMachineVolumeExport
	State VirtualMachineVolumeExportState `json:"state"`
	// Conditions indicate current conditions of VirtualMachineVolumeExport
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`
}

const (
	// VirtualMachineVolumeExportConditionReadyToUse indicated vmvExport is ready to use
	VirtualMachineVolumeExportConditionReadyToUse = "ReadyToUse"
)

// VirtualMachineVolumeExportState is the current state of the VirtualMachineVolumeExport
type VirtualMachineVolumeExportState string

const (
	// VirtualMachineVolumeExportStateCreating indicates VirtualMachineVolumeExport is creating
	VirtualMachineVolumeExportStateCreating VirtualMachineVolumeExportState = "Creating"
	// VirtualMachineVolumeExportStateCompleted indicates the pvc export is completed
	VirtualMachineVolumeExportStateCompleted VirtualMachineVolumeExportState = "Completed"
	// VirtualMachineVolumeExportStateError indicates VirtualMachineVolumeExport is error
	VirtualMachineVolumeExportStateError VirtualMachineVolumeExportState = "Error"
	// VirtualMachineVolumeExportStatePending indicates VirtualMachineVolumeExport is pending
	VirtualMachineVolumeExportStatePending VirtualMachineVolumeExportState = "Pending"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualMachineVolumeExport is the Schema for the virtualmachinevolumeexports API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=virtualmachinevolumeexports,scope=Namespaced,shortName=vmve
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="Current state of VirtualMachineVolumeExport"
type VirtualMachineVolumeExport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtualMachineVolumeExportSpec   `json:"spec,omitempty"`
	Status VirtualMachineVolumeExportStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualMachineVolumeExportList contains a list of VirtualMachineVolumeExport
type VirtualMachineVolumeExportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualMachineVolumeExport `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtualMachineVolumeExport{}, &VirtualMachineVolumeExportList{})
}
