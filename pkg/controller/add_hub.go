package controller

import (
	"github.com/sm43/hub-operator/pkg/controller/hub"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, hub.Add)
}
