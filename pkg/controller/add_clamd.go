package controller

import (
	"github.com/rhdedgar/scanning-operator/pkg/controller/clamd"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, clamd.Add)
}
