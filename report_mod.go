package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/ds"
	"github.com/roidaradal/fn/io"
	"github.com/roidaradal/fn/lang"
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
		addStatsReport,
		addDepsReport,
		addCodeReport,
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
	var key string
	table := make([]string, 0)
	fileCounts := list.Map(mod.Packages, (*Package).FileCount)
	pkgFileCounts := dict.Entries(dict.Zip(mod.PackageNames(), fileCounts))
	slices.SortFunc(pkgFileCounts, sortDescCount)
	lookup := ds.NewLookupCode(mod.Packages)
	hasTest := mod.Stats.Files[FILE_TEST] > 0
	for _, e := range pkgFileCounts {
		pkgName, count := e.Tuple()
		pkg := lookup[pkgName]
		node := mod.Nodes[packageToNodeName(pkgName)]
		table = append(table,
			"<tr>",
			wrapTdTags(pkgName, ""),
			wrapTdTags(percentage(count, mod.Stats.FileCount), "center"),
			wrapTdTags(number.Comma(count), "center"),
			lang.Ternary(hasTest,
				wrapTdTags(fmt.Sprintf("%d | %d", pkg.FileTypes[FILE_CODE], pkg.FileTypes[FILE_TEST]), "center"),
				"",
			),
			wrapTdTags(strings.Join(node.Files, "<br/>"), ""),
			"</tr>",
		)
	}
	rep["ModPackageCount"] = number.Comma(mod.Stats.PackageCount)
	for _, pkgType := range []PackageType{PKG_LIB, PKG_MAIN} {
		key = fmt.Sprintf("%sPackageCount", pkgType)
		rep[key] = number.Comma(mod.Stats.Packages[pkgType])
	}

	fileCount := mod.Stats.FileCount
	rep["ModFileCount"] = number.Comma(fileCount)
	for _, fileType := range []FileType{FILE_CODE, FILE_TEST} {
		typeCount := mod.Stats.Files[fileType]
		key = fmt.Sprintf("%sFileCount", fileType)
		rep[key] = number.Comma(typeCount)

		key = fmt.Sprintf("%sFileShare", fileType)
		rep[key] = percentage(typeCount, fileCount)
	}
	rep["ModuleTableHeader"] = lang.Ternary(hasTest, "<th>Split</th>", "")
	rep["ModuleTable"] = strings.Join(table, "")
}

// Create the template key: %key%
func templateKey(key string) string {
	return fmt.Sprintf("%%%s%%", key)
}
