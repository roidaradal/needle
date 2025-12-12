package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/ds"
	"github.com/roidaradal/fn/list"
	"github.com/roidaradal/fn/number"
)

var lineTypes = []LineType{LINE_CODE, LINE_ERROR, LINE_HEAD, LINE_COMMENT, LINE_SPACE}

// Add code report data
func addCodeReport(mod *Module, rep dict.StringMap) {
	var key string

	// Code Header
	modLineCount := mod.Stats.LineCount
	modCharCount := mod.Stats.CharCount
	for _, lineType := range lineTypes {
		lineCount := mod.Code.Lines[lineType]
		charCount := mod.Code.Chars[lineType]

		key = fmt.Sprintf("%sLineCount", lineType)
		rep[key] = number.Comma(lineCount)

		key = fmt.Sprintf("%sLineShare", lineType)
		rep[key] = percentage(lineCount, modLineCount)

		key = fmt.Sprintf("%sCharCount", lineType)
		rep[key] = number.Comma(charCount)

		key = fmt.Sprintf("%sCharShare", lineType)
		rep[key] = percentage(charCount, modCharCount)
	}

	members := map[BlockType][2][]CodeType{
		CODE_GLOBAL: {
			{PUB_CONST, PUB_VAR},
			{PRIV_CONST, PRIV_VAR},
		},
		CODE_FUNCTION: {
			{PUB_FUNCTION, PUB_METHOD},
			{PRIV_FUNCTION, PRIV_METHOD},
		},
		CODE_TYPE: {
			{PUB_STRUCT, PUB_INTERFACE, PUB_ALIAS},
			{PRIV_STRUCT, PRIV_INTERFACE, PRIV_ALIAS},
		},
	}
	visibility := []string{"Public", "Private"}
	for _, blockType := range []BlockType{CODE_GLOBAL, CODE_FUNCTION, CODE_TYPE} {
		key = fmt.Sprintf("%sCount", blockType)
		rep[key] = number.Comma(mod.Code.Blocks[blockType])

		for i, vis := range visibility {
			key = fmt.Sprintf("%s%sCount", vis, blockType)
			rep[key] = number.Comma(list.Sum(list.Translate(members[blockType][i], mod.Code.Types)))
		}
	}

	lookup := ds.NewLookupCode(mod.Packages)

	rep["CodeLinesTable"] = newCodeBreakdown(mod, &codeBreakdownConfig{
		name:       "lines",
		lookup:     lookup,
		modCount:   mod.Stats.LineCount,
		modCounter: mod.Code.Lines,
		pkgCounter: func(pkg *Package) dict.Counter[LineType] {
			return pkg.LineTypes
		},
		countFn: func(pkg *Package) int {
			return pkg.LineCount
		},
	})

	rep["CodeCharsTable"] = newCodeBreakdown(mod, &codeBreakdownConfig{
		name:       "chars",
		lookup:     lookup,
		modCount:   mod.Stats.CharCount,
		modCounter: mod.Code.Chars,
		pkgCounter: func(pkg *Package) dict.Counter[LineType] {
			return pkg.CharTypes
		},
		countFn: func(pkg *Package) int {
			return pkg.CharCount
		},
	})

	rep["GlobalsTable"] = newCodeTable(mod, &codeConfig{
		name:      "globals",
		lookup:    lookup,
		blockType: CODE_GLOBAL,
		keys:      []CodeType{PUB_CONST, PRIV_CONST, PUB_VAR, PRIV_VAR},
	})

	rep["FunctionsTable"] = newCodeTable(mod, &codeConfig{
		name:      "functions",
		lookup:    lookup,
		blockType: CODE_FUNCTION,
		keys:      []CodeType{PUB_FUNCTION, PRIV_FUNCTION, PUB_METHOD, PRIV_METHOD},
	})

	rep["TypesTable"] = newCodeTable(mod, &codeConfig{
		name:      "types",
		lookup:    lookup,
		blockType: CODE_TYPE,
		keys:      []CodeType{PUB_STRUCT, PRIV_STRUCT, PUB_INTERFACE, PRIV_INTERFACE, PUB_ALIAS, PRIV_ALIAS},
	})
}

type codeBreakdownConfig struct {
	name       string
	lookup     ds.LookupCode[*Package]
	modCount   int
	modCounter dict.Counter[LineType]
	pkgCounter func(*Package) dict.Counter[LineType]
	countFn    func(*Package) int
}

// Create new code breakdown table (lines / chars)
func newCodeBreakdown(mod *Module, cfg *codeBreakdownConfig) string {
	detailsClass := fmt.Sprintf(" hidden code-%s-list", cfg.name)
	table := make([]string, 0)
	counts := list.Map(mod.Packages, cfg.countFn)
	pkgEntries := dict.Entries(dict.Zip(mod.PackageNames(), counts))
	slices.SortFunc(pkgEntries, sortDescCount)
	for _, e := range pkgEntries {
		pkgName, pkgCount := e.Tuple()
		pkg := cfg.lookup[pkgName]
		counter := cfg.pkgCounter(pkg)
		table = append(table,
			"<tr>",
			wrapTag(td, pkgName, withRowspan(2)),
			wrapTag(td, number.Comma(pkgCount), withClass(centerLocal)),
		)
		row := []string{
			"</tr><tr>",
			wrapTag(td, percentage(pkgCount, cfg.modCount), withClass(center+detailsClass)),
		}
		for _, lineType := range lineTypes {
			typeCount := counter[lineType]
			table = append(table, wrapTag(td, number.Comma(typeCount), withClass(center), withColspan(2)))
			row = append(row,
				wrapTag(td, percentage(typeCount, pkgCount), withClass(centerLocal+detailsClass)),
				wrapTag(td, percentage(typeCount, cfg.modCounter[lineType]), withClass(centerGlobal+detailsClass)),
			)
		}
		table = append(table, strings.Join(row, ""), "</tr>")
	}
	// Footer
	table = append(table,
		"<tr>",
		wrapTag(td, "TOTAL", withClass(center), withRowspan(2)),
		wrapTag(td, number.Comma(cfg.modCount), withClass(center), withRowspan(2)),
	)
	row := []string{"</tr><tr>"}
	for _, lineType := range lineTypes {
		typeCount := cfg.modCounter[lineType]
		bottomCell := fmt.Sprintf("%s<br/><b>%s</b>", percentage(typeCount, cfg.modCount), lineType)

		table = append(table, wrapTag(td, number.Comma(typeCount), withClass(centerGlobal), withColspan(2)))
		row = append(row, wrapTag(td, bottomCell, withClass(center), withColspan(2)))
	}
	table = append(table, strings.Join(row, ""), "</tr>")
	return strings.Join(table, "")
}

type codeConfig struct {
	name      string
	lookup    ds.LookupCode[*Package]
	blockType BlockType
	keys      []CodeType
}

// Create new code table (globals / functions / types)
func newCodeTable(mod *Module, cfg *codeConfig) string {
	detailsClass := fmt.Sprintf(" hidden code-%s-list", cfg.name)
	table := make([]string, 0)
	activeKeys := make([]CodeType, 0)

	// Header
	colspan := 2
	table = append(table,
		"<thead><tr>",
		wrapTag(th, "Packages"),
		wrapTag(th, string(cfg.blockType)+"s"),
	)
	for _, key := range cfg.keys {
		if mod.Code.Types[key] == 0 {
			continue
		}
		activeKeys = append(activeKeys, key)

		table = append(table, wrapTag(th, string(key), withColspan(2)))
		colspan += 2
	}
	onclickFn := fmt.Sprintf("toggleList('code','%s','%%')", cfg.name)
	table = append(table, wrapTag(th, button("Show %", withID("toggle-code-"+cfg.name), onclick(onclickFn))))
	table = append(table, "</tr></thead><tbody>")

	// Body
	modCount := mod.Code.Blocks[cfg.blockType]
	blankPackages := make([]string, 0)

	counts := list.Map(mod.Packages, func(pkg *Package) int {
		return pkg.Blocks[cfg.blockType]
	})
	pkgEntries := dict.Entries(dict.Zip(mod.PackageNames(), counts))
	slices.SortFunc(pkgEntries, sortDescCount)
	for _, e := range pkgEntries {
		pkgName, pkgCount := e.Tuple()
		if pkgCount == 0 {
			blankPackages = append(blankPackages, pkgName)
			continue
		}
		pkg := cfg.lookup[pkgName]
		table = append(table,
			"<tr>",
			wrapTag(td, pkgName, withRowspan(2)),
			wrapTag(td, number.Comma(pkgCount), withClass(centerLocal)),
		)
		row := []string{
			"</tr><tr>",
			wrapTag(td, percentage(pkgCount, modCount), withClass(center+detailsClass)),
		}
		for _, key := range activeKeys {
			count := pkg.Codes[key]
			table = append(table, wrapTag(td, number.Comma(count), withClass(center), withColspan(2)))
			row = append(row,
				wrapTag(td, percentage(count, pkgCount), withClass(centerLocal+detailsClass)),
				wrapTag(td, percentage(count, mod.Code.Types[key]), withClass(centerGlobal+detailsClass)),
			)
		}
		table = append(table, strings.Join(row, ""), "</tr>")
	}

	// Footer
	table = append(table,
		"<tr>",
		wrapTag(td, "TOTAL", withClass(center), withRowspan(2)),
		wrapTag(td, number.Comma(modCount), withClass(center), withRowspan(2)),
	)
	row := []string{"</tr><tr>"}
	for _, key := range activeKeys {
		count := mod.Code.Types[key]
		bottomCell := fmt.Sprintf("%s<br/><b>%s</b>", percentage(count, modCount), key)
		table = append(table, wrapTag(td, number.Comma(count), withClass(centerGlobal), withColspan(2)))
		row = append(row, wrapTag(td, bottomCell, withClass(center), withColspan(2)))
	}

	var lastRow string = ""
	if len(blankPackages) > 0 {
		lastRow = fmt.Sprintf("Packages without %ss: %d<br/>%s", cfg.blockType, len(blankPackages), strings.Join(blankPackages, ", "))
		lastRow = "<tr>" + wrapTag(td, lastRow, withColspan(colspan)) + "</tr>"
	}
	table = append(table,
		strings.Join(row, ""), "</tr>",
		lastRow,
		"</tbody>",
	)
	return strings.Join(table, "")
}
