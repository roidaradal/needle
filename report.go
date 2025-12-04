package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/io"
	"github.com/roidaradal/fn/list"
	"github.com/roidaradal/fn/number"
)

// Create report HTML file from template
func BuildReport(mod *Module, path string) error {
	// TODO: Load template from string variable instead of file
	report, err := io.ReadFile("template.html")
	if err != nil {
		return err
	}
	// Initialize replacements
	replacements := dict.StringMap{
		"ModuleName": mod.Name,
	}
	// Apply report decorators
	decorators := []func(*Module, dict.StringMap){
		addModReport,
		addDepsReport,
	}
	for _, decorator := range decorators {
		decorator(mod, replacements)
	}
	// Replace template placeholders
	for key, replacement := range replacements {
		key = templateKey(key)
		report = strings.ReplaceAll(report, key, replacement)
	}
	// Save report to output file
	return io.SaveString(report, path)
}

// Add module report data
func addModReport(mod *Module, rep dict.StringMap) {
	names := list.Filter(dict.Keys(mod.Nodes), func(name string) bool {
		return mod.Nodes[name].FileCount() > 0
	})
	slices.Sort(names)
	modTree := make([]string, 0)
	for _, name := range names {
		node := mod.Nodes[name]
		modTree = append(modTree,
			fmt.Sprintf("<li>(%d) %s<ul>", node.FileCount(), name),
			listItems(node.Files),
			"</ul></li>",
		)
	}
	rep["ModPackageCount"] = number.Comma(mod.Stats.PackageCount)
	rep["ModFileCount"] = number.Comma(mod.Stats.FileCount)
	rep["ModuleTree"] = strings.Join(modTree, "")
}

// Create the template key: %key%
func templateKey(key string) string {
	return fmt.Sprintf("%%%s%%", key)
}
