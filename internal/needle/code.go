package needle

import (
	"fmt"
	"slices"
	"strings"

	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/ds"
	"github.com/roidaradal/fn/io"
	"github.com/roidaradal/fn/list"
	"github.com/roidaradal/fn/number"
	"github.com/roidaradal/fn/str"
)

const (
	modeNone = iota
	modeComment
	modeHead
	modeError
)

const (
	errLine   string = "if err != nil {"
	errSuffix string = "; if err != nil {"
)

var divider = str.Repeat(75, "-", "")

// Build new CodeModule for Go module at path
func NewCodeModule(cfg *Config) (*CodeModule, error) {
	// Initialize code module
	baseMod, err := baseModule(cfg)
	if err != nil {
		return nil, err
	}
	mod := newCodeModule(baseMod)

	// Run concurrently
	taskCfg := &taskConfig[TreeEntry, *Package]{
		Task: func(entry TreeEntry) (*Package, error) {
			folder, node := entry.Tuple()
			return newPackage(mod.Module, folder, node.Files, newCodeFile)
		},
		Receive: func(pkg *Package) {
			mod.Packages = append(mod.Packages, pkg)
		},
	}
	entries := mod.ValidTreeEntries()
	err = runConcurrent(entries, taskCfg)
	if err != nil {
		return nil, err
	}
	return mod, nil
}

// Build File object for CodeModule from file path
func newCodeFile(pkg *Package, path string) (*File, error) {
	file := &File{
		Name: getFilename(path),
		Type: getFileType(path),
	}
	lines, err := io.ReadRawLines(path)
	if err != nil {
		return nil, err
	}

	var errCloser string
	currMode := modeNone
	for _, rawLine := range lines {
		var line *Line
		cleanLine := strings.TrimSpace(rawLine)
		rawCount := len(rawLine)
		if cleanLine == "" {
			// Whitespace
			line = newSpaceLine()
		} else if strings.HasPrefix(cleanLine, "// ") {
			// Single-Line comment
			line = newCommentLine(rawCount)
		} else if strings.HasPrefix(cleanLine, "/*") {
			// Start Multi-Line comment
			line = newCommentLine(rawCount)
			currMode = modeComment
		} else if currMode == modeComment {
			// Inside Multi-Line comment
			line = newCommentLine(rawCount)
			// Close Multi-Line comment
			if strings.HasSuffix(cleanLine, "*/") {
				currMode = modeNone
			}
		} else if strings.HasPrefix(cleanLine, "package ") {
			// Package header
			line = newHeadLine(rawCount)
		} else if strings.HasPrefix(cleanLine, "import ") {
			// Import header
			line = newHeadLine(rawCount)
			if strings.HasSuffix(cleanLine, " (") {
				currMode = modeHead
			}
		} else if currMode == modeHead {
			// Inside Header mode
			line = newHeadLine(rawCount)
			// Close Header mode
			if cleanLine == ")" {
				currMode = modeNone
			}
		} else if cleanLine == errLine || strings.HasSuffix(cleanLine, errSuffix) {
			// Start error mode
			line = newErrorLine(rawCount)
			currMode = modeError
			errCloser = getIndentation(rawLine) + "}" + getTrailingWhitespace(rawLine)
		} else if currMode == modeError {
			// Inside Error mode
			line = newErrorLine(rawCount)
			// Close Error mode
			if rawLine == errCloser {
				currMode = modeNone
			}
		} else {
			// Default: normal code line
			line = newCodeLine(rawCount)
		}
		file.Lines = append(file.Lines, line)
	}

	return file, nil
}

// Create new Line with type: LINE_SPACE
func newSpaceLine() *Line {
	return &Line{Type: LINE_SPACE, Length: 1}
}

// Create new Line with type: LINE_COMMENT
func newCommentLine(length int) *Line {
	return &Line{Type: LINE_COMMENT, Length: length}
}

// Create new Line with type: LINE_HEAD
func newHeadLine(length int) *Line {
	return &Line{Type: LINE_HEAD, Length: length}
}

// Create new Line with type: LINE_ERROR
func newErrorLine(length int) *Line {
	return &Line{Type: LINE_ERROR, Length: length}
}

// Create new Line with type: LINE_CODE
func newCodeLine(length int) *Line {
	return &Line{Type: LINE_CODE, Length: length}
}

// CodeModule string representation
func (mod CodeModule) String() string {
	out := []string{
		mod.Header(),
		mod.PackageLineBreakdown(),
		mod.PackageCharBreakdown(),
	}
	return strings.Join(out, "\n")
}

// Display CodeModule header
func (mod CodeModule) Header() string {
	out := []string{
		fmt.Sprintf("Name: %s", mod.Name),
		fmt.Sprintf("Lines: %s", number.Comma(mod.CountLines())),
		fmt.Sprintf("Chars: %s", number.Comma(mod.CountChars())),
		divider,
	}
	return strings.Join(out, "\n")
}

// Per package: breakdown line usage
func (mod CodeModule) PackageLineBreakdown() string {
	out := []string{
		"LINE BREAKDOWN",
		divider,
		breakdownHeader(),
		divider,
	}
	ratioTmpl := "%s,%s"

	// Sort packages by total lines
	modUsage := mod.BreakdownLines()
	lookup := ds.NewLookupCode(mod.Packages)
	pkgNames := mod.PackageNames()
	lineCounts := list.Map(mod.Packages, (*Package).CountLines)
	pkgLineCounts := dict.Entries(dict.Zip(pkgNames, lineCounts))
	slices.SortFunc(pkgLineCounts, sortDescCount)
	totalLines := list.Sum(lineCounts)
	for _, e := range pkgLineCounts {
		pkgName, pkgLineCount := e.Tuple()
		pkg := lookup[pkgName]
		usage := pkg.BreakdownLines()
		out = append(out, breakdownRow(
			usage[LINE_CODE],
			usage[LINE_ERROR],
			usage[LINE_HEAD],
			usage[LINE_COMMENT],
			usage[LINE_SPACE],
			pkgLineCount,
			pkgName,
		))
		if mod.ShowDetails {
			out = append(out, breakdownRowString(
				fmt.Sprintf(ratioTmpl, percentage(usage[LINE_CODE], modUsage[LINE_CODE]), percentage(usage[LINE_CODE], pkgLineCount)),
				fmt.Sprintf(ratioTmpl, percentage(usage[LINE_ERROR], modUsage[LINE_ERROR]), percentage(usage[LINE_ERROR], pkgLineCount)),
				fmt.Sprintf(ratioTmpl, percentage(usage[LINE_HEAD], modUsage[LINE_HEAD]), percentage(usage[LINE_HEAD], pkgLineCount)),
				fmt.Sprintf(ratioTmpl, percentage(usage[LINE_COMMENT], modUsage[LINE_COMMENT]), percentage(usage[LINE_COMMENT], pkgLineCount)),
				fmt.Sprintf(ratioTmpl, percentage(usage[LINE_SPACE], modUsage[LINE_SPACE]), percentage(usage[LINE_SPACE], pkgLineCount)),
				percentage(pkgLineCount, totalLines), "",
			))
		}
		out = append(out, divider)
	}

	out = append(out, breakdownRow(
		modUsage[LINE_CODE],
		modUsage[LINE_ERROR],
		modUsage[LINE_HEAD],
		modUsage[LINE_COMMENT],
		modUsage[LINE_SPACE],
		totalLines,
		"TOTAL",
	))
	out = append(out, breakdownRowString(
		percentage(modUsage[LINE_CODE], totalLines),
		percentage(modUsage[LINE_ERROR], totalLines),
		percentage(modUsage[LINE_HEAD], totalLines),
		percentage(modUsage[LINE_COMMENT], totalLines),
		percentage(modUsage[LINE_SPACE], totalLines),
		"", "",
	))
	out = append(out, divider)

	return strings.Join(out, "\n")
}

// Per package: breakdown char usage
func (mod CodeModule) PackageCharBreakdown() string {
	out := []string{
		"CHAR BREAKDOWN",
		divider,
		breakdownHeader(),
		divider,
	}
	ratioTmpl := "%s,%s"

	// Sort packages by total chars
	modUsage := mod.BreakdownChars()
	lookup := ds.NewLookupCode(mod.Packages)
	pkgNames := mod.PackageNames()
	charCounts := list.Map(mod.Packages, (*Package).CountChars)
	pkgCharCounts := dict.Entries(dict.Zip(pkgNames, charCounts))
	slices.SortFunc(pkgCharCounts, sortDescCount)
	totalChars := list.Sum(charCounts)
	for _, e := range pkgCharCounts {
		pkgName, pkgCharCount := e.Tuple()
		pkg := lookup[pkgName]
		usage := pkg.BreakdownChars()
		out = append(out, breakdownRow(
			usage[LINE_CODE],
			usage[LINE_ERROR],
			usage[LINE_HEAD],
			usage[LINE_COMMENT],
			usage[LINE_SPACE],
			pkgCharCount,
			pkgName,
		))
		if mod.ShowDetails {
			out = append(out, breakdownRowString(
				fmt.Sprintf(ratioTmpl, percentage(usage[LINE_CODE], modUsage[LINE_CODE]), percentage(usage[LINE_CODE], pkgCharCount)),
				fmt.Sprintf(ratioTmpl, percentage(usage[LINE_ERROR], modUsage[LINE_ERROR]), percentage(usage[LINE_ERROR], pkgCharCount)),
				fmt.Sprintf(ratioTmpl, percentage(usage[LINE_HEAD], modUsage[LINE_HEAD]), percentage(usage[LINE_HEAD], pkgCharCount)),
				fmt.Sprintf(ratioTmpl, percentage(usage[LINE_COMMENT], modUsage[LINE_COMMENT]), percentage(usage[LINE_COMMENT], pkgCharCount)),
				fmt.Sprintf(ratioTmpl, percentage(usage[LINE_SPACE], modUsage[LINE_SPACE]), percentage(usage[LINE_SPACE], pkgCharCount)),
				percentage(pkgCharCount, totalChars), "",
			))
		}
		out = append(out, divider)
	}

	out = append(out, breakdownRow(
		modUsage[LINE_CODE],
		modUsage[LINE_ERROR],
		modUsage[LINE_HEAD],
		modUsage[LINE_COMMENT],
		modUsage[LINE_SPACE],
		totalChars,
		"TOTAL",
	))
	out = append(out, breakdownRowString(
		percentage(modUsage[LINE_CODE], totalChars),
		percentage(modUsage[LINE_ERROR], totalChars),
		percentage(modUsage[LINE_HEAD], totalChars),
		percentage(modUsage[LINE_COMMENT], totalChars),
		percentage(modUsage[LINE_SPACE], totalChars),
		"", "",
	))
	out = append(out, divider)

	return strings.Join(out, "\n")
}

// Create breakdown header
func breakdownHeader() string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
		strCenter("Code", 8),
		strCenter("Error", 8),
		strCenter("Head", 8),
		strCenter("Comment", 8),
		strCenter("Space", 8),
		strCenter("Total", 10),
		strCenter("Packages", 20),
	)
}

// Create breakdown row
func breakdownRow(code, err, head, comment, space, total int, pkg string) string {
	return fmt.Sprintf("%8s|%8s|%8s|%8s|%8s|%10s| %s",
		number.Comma(code),
		number.Comma(err),
		number.Comma(head),
		number.Comma(comment),
		number.Comma(space),
		number.Comma(total),
		pkg,
	)
}

// Create breakdown percentage with 1 value
func breakdownRowString(code, err, head, comment, space, total, pkg string) string {
	return fmt.Sprintf("%8s|%8s|%8s|%8s|%8s|%10s| %s",
		code,
		err,
		head,
		comment,
		space,
		total,
		pkg,
	)
}
