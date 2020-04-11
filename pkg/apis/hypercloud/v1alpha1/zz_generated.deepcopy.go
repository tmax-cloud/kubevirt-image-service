// +build !ignore_autogenerated

// Code generated by operator-sdk. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VirtualMachineImage) DeepCopyInto(out *VirtualMachineImage) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VirtualMachineImage.
func (in *VirtualMachineImage) DeepCopy() *VirtualMachineImage {
	if in == nil {
		return nil
	}
	out := new(VirtualMachineImage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VirtualMachineImage) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VirtualMachineImageList) DeepCopyInto(out *VirtualMachineImageList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VirtualMachineImage, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VirtualMachineImageList.
func (in *VirtualMachineImageList) DeepCopy() *VirtualMachineImageList {
	if in == nil {
		return nil
	}
	out := new(VirtualMachineImageList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VirtualMachineImageList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VirtualMachineImageSource) DeepCopyInto(out *VirtualMachineImageSource) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VirtualMachineImageSource.
func (in *VirtualMachineImageSource) DeepCopy() *VirtualMachineImageSource {
	if in == nil {
		return nil
	}
	out := new(VirtualMachineImageSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VirtualMachineImageSpec) DeepCopyInto(out *VirtualMachineImageSpec) {
	*out = *in
	out.Source = in.Source
	in.PVC.DeepCopyInto(&out.PVC)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VirtualMachineImageSpec.
func (in *VirtualMachineImageSpec) DeepCopy() *VirtualMachineImageSpec {
	if in == nil {
		return nil
	}
	out := new(VirtualMachineImageSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VirtualMachineImageStatus) DeepCopyInto(out *VirtualMachineImageStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VirtualMachineImageStatus.
func (in *VirtualMachineImageStatus) DeepCopy() *VirtualMachineImageStatus {
	if in == nil {
		return nil
	}
	out := new(VirtualMachineImageStatus)
	in.DeepCopyInto(out)
	return out
}
