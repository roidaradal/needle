package main

import (
	"slices"
	"strings"

	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/ds"
	"github.com/roidaradal/fn/list"
	"github.com/roidaradal/fn/number"
)

// Add code report data
func addCodeReport(mod *Module, rep dict.StringMap) {
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
		table = append(table,
			"<tr>",
			wrapTdTagsRowspan(pkgName, "", 2),
			wrapTdTagsRowspan(number.Comma(pkgLineCount), "center local", 2),
			wrapTdTagsColspan(number.Comma(pkg.LineTypes[LINE_CODE]), "center", 2),
			wrapTdTagsColspan(number.Comma(pkg.LineTypes[LINE_ERROR]), "center", 2),
			wrapTdTagsColspan(number.Comma(pkg.LineTypes[LINE_HEAD]), "center", 2),
			wrapTdTagsColspan(number.Comma(pkg.LineTypes[LINE_COMMENT]), "center", 2),
			wrapTdTagsColspan(number.Comma(pkg.LineTypes[LINE_SPACE]), "center", 2),
			"</tr><tr>",
			wrapTdTags(percentage(pkg.LineTypes[LINE_CODE], pkgLineCount), "center local"),
			wrapTdTags(percentage(pkg.LineTypes[LINE_CODE], mod.Code.Lines[LINE_CODE]), "center global"),
			wrapTdTags(percentage(pkg.LineTypes[LINE_ERROR], pkgLineCount), "center local"),
			wrapTdTags(percentage(pkg.LineTypes[LINE_ERROR], mod.Code.Lines[LINE_ERROR]), "center global"),
			wrapTdTags(percentage(pkg.LineTypes[LINE_HEAD], pkgLineCount), "center local"),
			wrapTdTags(percentage(pkg.LineTypes[LINE_HEAD], mod.Code.Lines[LINE_HEAD]), "center global"),
			wrapTdTags(percentage(pkg.LineTypes[LINE_COMMENT], pkgLineCount), "center local"),
			wrapTdTags(percentage(pkg.LineTypes[LINE_COMMENT], mod.Code.Lines[LINE_COMMENT]), "center global"),
			wrapTdTags(percentage(pkg.LineTypes[LINE_SPACE], pkgLineCount), "center local"),
			wrapTdTags(percentage(pkg.LineTypes[LINE_SPACE], mod.Code.Lines[LINE_SPACE]), "center global"),
			"</tr>",
		)
	}
	table = append(table,
		"<tr>",
		wrapTdTagsRowspan("TOTAL", "center", 2),
		wrapTdTagsRowspan(number.Comma(mod.Stats.LineCount), "center", 2),
		wrapTdTagsColspan(number.Comma(mod.Code.Lines[LINE_CODE]), "center", 2),
		wrapTdTagsColspan(number.Comma(mod.Code.Lines[LINE_ERROR]), "center", 2),
		wrapTdTagsColspan(number.Comma(mod.Code.Lines[LINE_HEAD]), "center", 2),
		wrapTdTagsColspan(number.Comma(mod.Code.Lines[LINE_COMMENT]), "center", 2),
		wrapTdTagsColspan(number.Comma(mod.Code.Lines[LINE_SPACE]), "center", 2),
		"</tr><tr>",
		wrapTdTagsColspan(percentage(mod.Code.Lines[LINE_CODE], mod.Stats.LineCount), "center global", 2),
		wrapTdTagsColspan(percentage(mod.Code.Lines[LINE_ERROR], mod.Stats.LineCount), "center global", 2),
		wrapTdTagsColspan(percentage(mod.Code.Lines[LINE_HEAD], mod.Stats.LineCount), "center global", 2),
		wrapTdTagsColspan(percentage(mod.Code.Lines[LINE_COMMENT], mod.Stats.LineCount), "center global", 2),
		wrapTdTagsColspan(percentage(mod.Code.Lines[LINE_SPACE], mod.Stats.LineCount), "center global", 2),
		"</tr>",
	)
	codeLinesTable := strings.Join(table, "")

	rep["CodesLineCount"] = number.Comma(mod.Code.Lines[LINE_CODE])
	rep["ErrorLineCount"] = number.Comma(mod.Code.Lines[LINE_ERROR])
	rep["HeadLineCount"] = number.Comma(mod.Code.Lines[LINE_HEAD])
	rep["CommentLineCount"] = number.Comma(mod.Code.Lines[LINE_COMMENT])
	rep["SpaceLineCount"] = number.Comma(mod.Code.Lines[LINE_SPACE])
	rep["CodesLineShare"] = percentage(mod.Code.Lines[LINE_CODE], mod.Stats.LineCount)
	rep["ErrorLineShare"] = percentage(mod.Code.Lines[LINE_ERROR], mod.Stats.LineCount)
	rep["HeadLineShare"] = percentage(mod.Code.Lines[LINE_HEAD], mod.Stats.LineCount)
	rep["CommentLineShare"] = percentage(mod.Code.Lines[LINE_COMMENT], mod.Stats.LineCount)
	rep["SpaceLineShare"] = percentage(mod.Code.Lines[LINE_SPACE], mod.Stats.LineCount)
	rep["CodeLinesTable"] = codeLinesTable

	// Characters table
	table = make([]string, 0)
	charCounts := list.Map(mod.Packages, func(pkg *Package) int {
		return pkg.CharCount
	})
	pkgCharCounts := dict.Entries(dict.Zip(pkgNames, charCounts))
	slices.SortFunc(pkgCharCounts, sortDescCount)
	for _, e := range pkgCharCounts {
		pkgName, pkgCharCount := e.Tuple()
		pkg := lookup[pkgName]
		table = append(table,
			"<tr>",
			wrapTdTagsRowspan(pkgName, "", 2),
			wrapTdTagsRowspan(number.Comma(pkgCharCount), "center local", 2),
			wrapTdTagsColspan(number.Comma(pkg.CharTypes[LINE_CODE]), "center", 2),
			wrapTdTagsColspan(number.Comma(pkg.CharTypes[LINE_ERROR]), "center", 2),
			wrapTdTagsColspan(number.Comma(pkg.CharTypes[LINE_HEAD]), "center", 2),
			wrapTdTagsColspan(number.Comma(pkg.CharTypes[LINE_COMMENT]), "center", 2),
			wrapTdTagsColspan(number.Comma(pkg.CharTypes[LINE_SPACE]), "center", 2),
			"</tr><tr>",
			wrapTdTags(percentage(pkg.CharTypes[LINE_CODE], pkgCharCount), "center local"),
			wrapTdTags(percentage(pkg.CharTypes[LINE_CODE], mod.Code.Chars[LINE_CODE]), "center global"),
			wrapTdTags(percentage(pkg.CharTypes[LINE_ERROR], pkgCharCount), "center local"),
			wrapTdTags(percentage(pkg.CharTypes[LINE_ERROR], mod.Code.Chars[LINE_ERROR]), "center global"),
			wrapTdTags(percentage(pkg.CharTypes[LINE_HEAD], pkgCharCount), "center local"),
			wrapTdTags(percentage(pkg.CharTypes[LINE_HEAD], mod.Code.Chars[LINE_HEAD]), "center global"),
			wrapTdTags(percentage(pkg.CharTypes[LINE_COMMENT], pkgCharCount), "center local"),
			wrapTdTags(percentage(pkg.CharTypes[LINE_COMMENT], mod.Code.Chars[LINE_COMMENT]), "center global"),
			wrapTdTags(percentage(pkg.CharTypes[LINE_SPACE], pkgCharCount), "center local"),
			wrapTdTags(percentage(pkg.CharTypes[LINE_SPACE], mod.Code.Chars[LINE_SPACE]), "center global"),
			"</tr>",
		)
	}
	table = append(table,
		"<tr>",
		wrapTdTagsRowspan("TOTAL", "center", 2),
		wrapTdTagsRowspan(number.Comma(mod.Stats.CharCount), "center", 2),
		wrapTdTagsColspan(number.Comma(mod.Code.Chars[LINE_CODE]), "center", 2),
		wrapTdTagsColspan(number.Comma(mod.Code.Chars[LINE_ERROR]), "center", 2),
		wrapTdTagsColspan(number.Comma(mod.Code.Chars[LINE_HEAD]), "center", 2),
		wrapTdTagsColspan(number.Comma(mod.Code.Chars[LINE_COMMENT]), "center", 2),
		wrapTdTagsColspan(number.Comma(mod.Code.Chars[LINE_SPACE]), "ricenterght", 2),
		"</tr><tr>",
		wrapTdTagsColspan(percentage(mod.Code.Chars[LINE_CODE], mod.Stats.CharCount), "center global", 2),
		wrapTdTagsColspan(percentage(mod.Code.Chars[LINE_ERROR], mod.Stats.CharCount), "center global", 2),
		wrapTdTagsColspan(percentage(mod.Code.Chars[LINE_HEAD], mod.Stats.CharCount), "center global", 2),
		wrapTdTagsColspan(percentage(mod.Code.Chars[LINE_COMMENT], mod.Stats.CharCount), "center global", 2),
		wrapTdTagsColspan(percentage(mod.Code.Chars[LINE_SPACE], mod.Stats.CharCount), "center global", 2),
		"</tr>",
	)
	codeCharsTable := strings.Join(table, "")

	rep["CodesCharCount"] = number.Comma(mod.Code.Chars[LINE_CODE])
	rep["ErrorCharCount"] = number.Comma(mod.Code.Chars[LINE_ERROR])
	rep["HeadCharCount"] = number.Comma(mod.Code.Chars[LINE_HEAD])
	rep["CommentCharCount"] = number.Comma(mod.Code.Chars[LINE_COMMENT])
	rep["SpaceCharCount"] = number.Comma(mod.Code.Chars[LINE_SPACE])
	rep["CodesCharShare"] = percentage(mod.Code.Chars[LINE_CODE], mod.Stats.CharCount)
	rep["ErrorCharShare"] = percentage(mod.Code.Chars[LINE_ERROR], mod.Stats.CharCount)
	rep["HeadCharShare"] = percentage(mod.Code.Chars[LINE_HEAD], mod.Stats.CharCount)
	rep["CommentCharShare"] = percentage(mod.Code.Chars[LINE_COMMENT], mod.Stats.CharCount)
	rep["SpaceCharShare"] = percentage(mod.Code.Chars[LINE_SPACE], mod.Stats.CharCount)
	rep["CodeCharsTable"] = codeCharsTable
}
