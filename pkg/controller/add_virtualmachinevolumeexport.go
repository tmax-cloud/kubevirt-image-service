package controller

import (
	"kubevirt-image-service/pkg/controller/virtualmachinevolumeexport"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, virtualmachinevolumeexport.Add)
}
