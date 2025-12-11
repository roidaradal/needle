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

var LineTypes = []LineType{LINE_CODE, LINE_ERROR, LINE_HEAD, LINE_COMMENT, LINE_SPACE}

// Add code report data
func addCodeReport(mod *Module, rep dict.StringMap) {
	var key string

	// Code Header
	modLineCount := mod.Stats.LineCount
	modCharCount := mod.Stats.CharCount
	for _, lineType := range LineTypes {
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
		lookup:    lookup,
		blockType: CODE_GLOBAL,
		keys:      []CodeType{PUB_CONST, PRIV_CONST, PUB_VAR, PRIV_VAR},
	})

	rep["FunctionsTable"] = newCodeTable(mod, &codeConfig{
		lookup:    lookup,
		blockType: CODE_FUNCTION,
		keys:      []CodeType{PUB_FUNCTION, PRIV_FUNCTION, PUB_METHOD, PRIV_METHOD},
	})

	rep["TypesTable"] = newCodeTable(mod, &codeConfig{
		lookup:    lookup,
		blockType: CODE_TYPE,
		keys:      []CodeType{PUB_STRUCT, PRIV_STRUCT, PUB_INTERFACE, PRIV_INTERFACE, PUB_ALIAS, PRIV_ALIAS},
	})
}

type codeBreakdownConfig struct {
	lookup     ds.LookupCode[*Package]
	modCount   int
	modCounter dict.Counter[LineType]
	pkgCounter func(*Package) dict.Counter[LineType]
	countFn    func(*Package) int
}

// Create new code breakdown table (lines / chars)
func newCodeBreakdown(mod *Module, cfg *codeBreakdownConfig) string {
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
			wrapTdTagsRowspan(pkgName, "", 2),
			wrapTdTags(number.Comma(pkgCount), "center local"),
		)
		row := []string{
			"</tr><tr>",
			wrapTdTags(percentage(pkgCount, cfg.modCount), "center"),
		}
		for _, lineType := range LineTypes {
			typeCount := counter[lineType]
			table = append(table, wrapTdTagsColspan(number.Comma(typeCount), "center", 2))
			row = append(row,
				wrapTdTags(percentage(typeCount, pkgCount), "center local"),
				wrapTdTags(percentage(typeCount, cfg.modCounter[lineType]), "center global"),
			)
		}
		table = append(table, strings.Join(row, ""), "</tr>")
	}
	// Footer
	table = append(table,
		"<tr>",
		wrapTdTagsRowspan("TOTAL", "center", 2),
		wrapTdTagsRowspan(number.Comma(cfg.modCount), "center", 2),
	)
	row := []string{"</tr><tr>"}
	for _, lineType := range LineTypes {
		typeCount := cfg.modCounter[lineType]
		table = append(table, wrapTdTagsColspan(number.Comma(typeCount), "center global", 2))
		row = append(row, wrapTdTagsColspan(percentage(typeCount, cfg.modCount), "center", 2))
	}
	table = append(table, strings.Join(row, ""), "</tr>")
	return strings.Join(table, "")
}

type codeConfig struct {
	lookup    ds.LookupCode[*Package]
	blockType BlockType
	keys      []CodeType
}

// Create new code table (globals / functions / types)
func newCodeTable(mod *Module, cfg *codeConfig) string {
	table := make([]string, 0)
	activeKeys := make([]CodeType, 0)

	// Header
	colspan := 2
	table = append(table,
		"<thead><tr>",
		"<th>Packages</th>",
		fmt.Sprintf("<th>%ss</th>", cfg.blockType),
	)
	for _, key := range cfg.keys {
		if mod.Code.Types[key] == 0 {
			continue
		}
		activeKeys = append(activeKeys, key)
		table = append(table, fmt.Sprintf("<th colspan='2'>%s</th>", key))
		colspan += 2
	}
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
			wrapTdTagsRowspan(pkgName, "", 2),
			wrapTdTags(number.Comma(pkgCount), "center local"),
		)
		row := []string{
			"</tr><tr>",
			wrapTdTags(percentage(pkgCount, modCount), "center"),
		}
		for _, key := range activeKeys {
			count := pkg.Codes[key]
			table = append(table, wrapTdTagsColspan(number.Comma(count), "center", 2))
			row = append(row,
				wrapTdTags(percentage(count, pkgCount), "center local"),
				wrapTdTags(percentage(count, mod.Code.Types[key]), "center global"),
			)
		}
		table = append(table, strings.Join(row, ""), "</tr>")
	}

	// Footer
	table = append(table,
		"<tr>",
		wrapTdTagsRowspan("TOTAL", "center", 2),
		wrapTdTagsRowspan(number.Comma(modCount), "center", 2),
	)
	row := []string{"</tr><tr>"}
	for _, key := range activeKeys {
		count := mod.Code.Types[key]
		table = append(table, wrapTdTagsColspan(number.Comma(count), "center global", 2))
		row = append(row, wrapTdTagsColspan(percentage(count, modCount), "center", 2))
	}

	var lastRow string = ""
	if len(blankPackages) > 0 {
		lastRow = fmt.Sprintf("Packages without %ss: %d<br/>%s", cfg.blockType, len(blankPackages), strings.Join(blankPackages, ", "))
		lastRow = "<tr>" + wrapTdTagsColspan(lastRow, "", colspan) + "</tr>"
	}
	table = append(table,
		strings.Join(row, ""), "</tr>",
		lastRow,
		"</tbody>",
	)
	return strings.Join(table, "")
}
