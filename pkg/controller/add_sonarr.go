package controller

import (
	"github.com/parflesh/sonarr-operator/pkg/controller/sonarr"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, sonarr.Add)
}
