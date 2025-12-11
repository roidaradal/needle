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

// Add stats report data
func addStatsReport(mod *Module, rep dict.StringMap) {
	var key string
	lookup := ds.NewLookupCode(mod.Packages)

	// Lines
	fileTypes := []FileType{FILE_CODE, FILE_TEST}
	fileCount := mod.Stats.FileCount
	lineCount := mod.Stats.LineCount
	rep["ModLineCount"] = number.Comma(lineCount)
	rep["AvgLinePerFile"] = average(lineCount, fileCount)
	for _, fileType := range fileTypes {
		typeCount := mod.Stats.FileLines[fileType]

		key = fmt.Sprintf("%sLineCount", fileType)
		rep[key] = number.Comma(typeCount)

		key = fmt.Sprintf("%sLineShare", fileType)
		rep[key] = percentage(typeCount, lineCount)

		key = fmt.Sprintf("%sALPF", fileType)
		rep[key] = average(typeCount, mod.Stats.Files[fileType])
	}
	rep["LinesTable"] = newStatsTable(mod, &statsConfig{
		lookup:   lookup,
		modCount: mod.Stats.LineCount,
		countFn: func(pkg *Package) int {
			return pkg.LineCount
		},
		fileFn: func(f *File) int {
			return len(f.Lines)
		},
		averageCols: func(pkgCount int, pkg *Package, rowspan int) []string {
			return []string{
				wrapTdTagsRowspan(average(pkgCount, pkg.FileCount()), "center", rowspan),
			}
		},
	})

	// Characters
	charCount := mod.Stats.CharCount
	rep["ModCharCount"] = number.Comma(charCount)
	rep["AvgCharPerFile"] = average(charCount, fileCount)
	rep["AvgCharPerLine"] = average(charCount, lineCount)
	for _, fileType := range fileTypes {
		typeCount := mod.Stats.FileChars[fileType]

		key = fmt.Sprintf("%sCharCount", fileType)
		rep[key] = number.Comma(typeCount)

		key = fmt.Sprintf("%sCharShare", fileType)
		rep[key] = percentage(typeCount, charCount)

		key = fmt.Sprintf("%sACPF", fileType)
		rep[key] = average(typeCount, mod.Stats.Files[fileType])

		key = fmt.Sprintf("%sACPL", fileType)
		rep[key] = average(typeCount, mod.Stats.FileLines[fileType])
	}
	rep["CharsTable"] = newStatsTable(mod, &statsConfig{
		lookup:   lookup,
		modCount: mod.Stats.CharCount,
		countFn: func(pkg *Package) int {
			return pkg.CharCount
		},
		fileFn: func(f *File) int {
			return f.CharCount
		},
		averageCols: func(pkgCount int, pkg *Package, rowspan int) []string {
			return []string{
				wrapTdTagsRowspan(average(pkgCount, pkg.FileCount()), "center", rowspan),
				wrapTdTagsRowspan(average(pkgCount, pkg.LineCount), "center", rowspan),
			}
		},
	})
}

type statsConfig struct {
	lookup      ds.LookupCode[*Package]
	modCount    int
	countFn     func(*Package) int
	fileFn      func(*File) int
	averageCols func(int, *Package, int) []string
}

// Create new stats table (lines / char)
func newStatsTable(mod *Module, cfg *statsConfig) string {
	table := make([]string, 0)
	counts := list.Map(mod.Packages, cfg.countFn)
	pkgEntries := dict.Entries(dict.Zip(mod.PackageNames(), counts))
	slices.SortFunc(pkgEntries, sortDescCount)
	for _, e := range pkgEntries {
		pkgName, pkgCount := e.Tuple()
		pkg := cfg.lookup[pkgName]
		rowspan := 1 + pkg.FileCount()
		table = append(table,
			"<tr>",
			wrapTdTagsRowspan(pkgName, "", rowspan),
			wrapTdTagsRowspan(percentage(pkgCount, cfg.modCount), "center", rowspan),
			wrapTdTagsRowspan(number.Comma(pkgCount), "center", rowspan),
		)
		table = append(table, cfg.averageCols(pkgCount, pkg, rowspan)...)
		table = append(table, "</tr>")

		fileCounts := list.Map(pkg.Files, cfg.fileFn)
		fileEntries := dict.Entries(dict.Zip(pkg.FileNames(), fileCounts))
		slices.SortFunc(fileEntries, sortDescCount)
		for _, e2 := range fileEntries {
			fileName, fileCount := e2.Tuple()
			table = append(table,
				"<tr>",
				wrapTdTags(number.Comma(fileCount), "right"),
				wrapTdTags(percentage(fileCount, pkgCount), "center"),
				wrapTdTags(fileName, ""),
				"</tr>",
			)
		}
	}

	return strings.Join(table, "")
}
