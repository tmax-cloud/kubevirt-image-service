package controller

import (
	"kubevirt-image-service/pkg/controller/virtualmachineimage"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, virtualmachineimage.Add)
}
