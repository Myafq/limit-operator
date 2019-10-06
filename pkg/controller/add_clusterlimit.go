package controller

import (
	"github.com/myafq/limit-operator/pkg/controller/clusterlimit"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, clusterlimit.Add)
}
