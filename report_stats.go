package main

import (
	"slices"
	"strings"

	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/ds"
	"github.com/roidaradal/fn/list"
	"github.com/roidaradal/fn/number"
)

// Add stats report data
func addStatsReport(mod *Module, rep dict.StringMap) {
	// Lines Table
	table := make([]string, 0)
	pkgNames := mod.PackageNames()
	lineCounts := list.Map(mod.Packages, func(pkg *Package) int {
		return pkg.LineCount
	})
	pkgLineCounts := dict.Entries(dict.Zip(pkgNames, lineCounts))
	slices.SortFunc(pkgLineCounts, sortDescCount)
	lookup := ds.NewLookupCode(mod.Packages)
	for _, e := range pkgLineCounts {
		pkgName, pkgLineCount := e.Tuple()
		pkg := lookup[pkgName]
		pkgFileCount := pkg.FileCount()
		rowspan := 1 + pkgFileCount
		table = append(table,
			"<tr>",
			wrapTdTagsSpan(pkgName, "", rowspan),
			wrapTdTagsSpan(percentage(pkgLineCount, mod.Stats.LineCount), "center", rowspan),
			wrapTdTagsSpan(number.Comma(pkgLineCount), "center", rowspan),
			wrapTdTagsSpan(average(pkgLineCount, pkgFileCount), "center", rowspan),
			"</tr>",
		)
		filenames := pkg.FileNames()
		lineCounts := list.Map(pkg.Files, (func(f *File) int {
			return len(f.Lines)
		}))
		fileLineCounts := dict.Entries(dict.Zip(filenames, lineCounts))
		slices.SortFunc(fileLineCounts, sortDescCount)
		for _, e2 := range fileLineCounts {
			filename, fileLineCount := e2.Tuple()
			table = append(table,
				"<tr>",
				wrapTdTags(number.Comma(fileLineCount), "right"),
				wrapTdTags(percentage(fileLineCount, pkgLineCount), "center"),
				wrapTdTags(filename, ""),
				"</tr>",
			)
		}
	}
	linesTable := strings.Join(table, "")

	// Chars Table
	table = make([]string, 0)
	charCounts := list.Map(mod.Packages, func(pkg *Package) int {
		return pkg.CharCount
	})
	pkgCharCounts := dict.Entries(dict.Zip(pkgNames, charCounts))
	slices.SortFunc(pkgCharCounts, sortDescCount)
	for _, e := range pkgCharCounts {
		pkgName, pkgCharCount := e.Tuple()
		pkg := lookup[pkgName]
		pkgFileCount := pkg.FileCount()
		rowspan := 1 + pkgFileCount
		table = append(table,
			"<tr>",
			wrapTdTagsSpan(pkgName, "", rowspan),
			wrapTdTagsSpan(percentage(pkgCharCount, mod.Stats.CharCount), "center", rowspan),
			wrapTdTagsSpan(number.Comma(pkgCharCount), "center", rowspan),
			wrapTdTagsSpan(average(pkgCharCount, pkgFileCount), "center", rowspan),
			wrapTdTagsSpan(average(pkgCharCount, pkg.LineCount), "center", rowspan),
			"</tr>",
		)
		filenames := pkg.FileNames()
		charCounts := list.Map(pkg.Files, func(f *File) int {
			return f.CharCount
		})
		fileCharCounts := dict.Entries(dict.Zip(filenames, charCounts))
		slices.SortFunc(fileCharCounts, sortDescCount)
		for _, e2 := range fileCharCounts {
			filename, fileCharCount := e2.Tuple()
			table = append(table,
				"<tr>",
				wrapTdTags(number.Comma(fileCharCount), "right"),
				wrapTdTags(percentage(fileCharCount, pkgCharCount), "center"),
				wrapTdTags(filename, ""),
				"</tr>",
			)
		}
	}
	charsTable := strings.Join(table, "")

	// Lines
	rep["ModLineCount"] = number.Comma(mod.Stats.LineCount)
	rep["CodeLineCount"] = number.Comma(mod.Stats.FileLines[FILE_CODE])
	rep["TestLineCount"] = number.Comma(mod.Stats.FileLines[FILE_TEST])
	rep["CodeLineShare"] = percentage(mod.Stats.FileLines[FILE_CODE], mod.Stats.LineCount)
	rep["TestLineShare"] = percentage(mod.Stats.FileLines[FILE_TEST], mod.Stats.LineCount)
	rep["AvgLinePerFile"] = average(mod.Stats.LineCount, mod.Stats.FileCount)
	rep["CodeALPF"] = average(mod.Stats.FileLines[FILE_CODE], mod.Stats.Files[FILE_CODE])
	rep["TestALPF"] = average(mod.Stats.FileLines[FILE_TEST], mod.Stats.Files[FILE_TEST])
	rep["LinesTable"] = linesTable

	// Characters
	rep["ModCharCount"] = number.Comma(mod.Stats.CharCount)
	rep["CodeCharCount"] = number.Comma(mod.Stats.FileChars[FILE_CODE])
	rep["TestCharCount"] = number.Comma(mod.Stats.FileChars[FILE_TEST])
	rep["CodeCharShare"] = percentage(mod.Stats.FileChars[FILE_CODE], mod.Stats.CharCount)
	rep["TestCharShare"] = percentage(mod.Stats.FileChars[FILE_TEST], mod.Stats.CharCount)
	rep["AvgCharPerFile"] = average(mod.Stats.CharCount, mod.Stats.FileCount)
	rep["CodeACPF"] = average(mod.Stats.FileChars[FILE_CODE], mod.Stats.Files[FILE_CODE])
	rep["TestACPF"] = average(mod.Stats.FileChars[FILE_TEST], mod.Stats.Files[FILE_TEST])
	rep["AvgCharPerLine"] = average(mod.Stats.CharCount, mod.Stats.LineCount)
	rep["CodeACPL"] = average(mod.Stats.FileChars[FILE_CODE], mod.Stats.FileLines[FILE_CODE])
	rep["TestACPL"] = average(mod.Stats.FileChars[FILE_TEST], mod.Stats.FileLines[FILE_TEST])
	rep["CharsTable"] = charsTable
}
