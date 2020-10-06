/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VirtualMachineImageSource represents the source for our VirtualMachineImage, this can be HTTP or host path
type VirtualMachineImageSource struct {
	HTTP     string                             `json:"http,omitempty"`
	HostPath *VirtualMachineImageSourceHostPath `json:"hostPath,omitempty"`
}

// VirtualMachineImageSourceHostPath provides the parameters to create a virtual machine image from a host path
type VirtualMachineImageSourceHostPath struct {
	Path     string `json:"path"`
	NodeName string `json:"nodeName"`
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

const (
	// ConditionReadyToUse indicated vmi is ready to use
	ConditionReadyToUse = "ReadyToUse"
)

// VirtualMachineImageConditionType defines the condition of VirtualMachineImage
type VirtualMachineImageConditionType string

// VirtualMachineImageSpec defines the desired state of VirtualMachineImage
type VirtualMachineImageSpec struct {
	Source            VirtualMachineImageSource        `json:"source"`
	PVC               corev1.PersistentVolumeClaimSpec `json:"pvc"`
	SnapshotClassName string                           `json:"snapshotClassName"`
}

// VirtualMachineImageStatus defines the observed state of VirtualMachineImage
type VirtualMachineImageStatus struct {
	// State is the current state of VirtualMachineImage
	State VirtualMachineImageState `json:"state"`
	// Conditions indicate current conditions of VirtualMachineImage
	// +optional
	apiextensions.CustomResourceDefinitionConditionType
	Conditions []Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=virtualmachineimages,scope=Namespaced,shortName=vmim
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="Current state of VirtualMachineImage"

// VirtualMachineImage is the Schema for the virtualmachineimages API
type VirtualMachineImage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtualMachineImageSpec   `json:"spec,omitempty"`
	Status VirtualMachineImageStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VirtualMachineImageList contains a list of VirtualMachineImage
type VirtualMachineImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualMachineImage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtualMachineImage{}, &VirtualMachineImageList{})
}
