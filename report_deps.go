package main

import (
	"fmt"
	"strings"

	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/str"
)

// Add dependency report data
func addDepsReport(mod *Module, rep dict.StringMap) {
	externalDepsCount := len(mod.Deps.ExternalUsers)
	externalDepsList := make([]string, 0, externalDepsCount)
	for extPkg := range mod.Deps.ExternalUsers {
		line := fmt.Sprintf("<li>%s</li>", extPkg)
		externalDepsList = append(externalDepsList, line)
	}
	rep["ExternalDepsCount"] = str.Int(externalDepsCount)
	rep["ExternalDepsList"] = strings.Join(externalDepsList, "")
}
