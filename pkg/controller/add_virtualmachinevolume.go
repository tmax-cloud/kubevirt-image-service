package controller

import (
	"kubevirt-image-service/pkg/controller/virtualmachinevolume"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, virtualmachinevolume.Add)
}
