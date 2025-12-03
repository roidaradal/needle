package needle

import (
	"fmt"
	"slices"
	"strings"

	"github.com/roidaradal/fn/conk"
	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/ds"
	"github.com/roidaradal/fn/io"
	"github.com/roidaradal/fn/lang"
	"github.com/roidaradal/fn/list"
	"github.com/roidaradal/fn/number"
	"github.com/roidaradal/fn/str"
)

const (
	modeNone = iota
	modeComment
	modeHead
	modeError
	modeFunction
	modeType
	modeTypeGroup
	modeVarGroup
	modeConstGroup
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
	task := func(entry TreeEntry) (*Package, error) {
		folder, node := entry.Tuple()
		return newPackage(mod.Module, folder, node.Files, newCodeFile)
	}
	onReceive := func(pkg *Package) {
		mod.Packages = append(mod.Packages, pkg)
	}
	entries := mod.ValidTreeEntries()
	err = conk.Tasks(entries, task, onReceive)
	if err != nil {
		return nil, err
	}
	return mod, nil
}

// Build File object for CodeModule from file path
func newCodeFile(pkg *Package, path string) (*File, error) {
	file := &File{
		Name:  getFilename(path),
		Type:  getFileType(path),
		Lines: make([]*Line, 0),
		Block: make(map[BlockType]int),
		Code:  make(map[CodeType]int),
	}
	lines, err := io.ReadRawLines(path)
	if err != nil {
		return nil, err
	}

	var modeCloser string
	var codeType CodeType
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
			// Set to comment mode if not a single-line block comment /*...*/
			if !strings.HasSuffix(cleanLine, "*/") {
				currMode = modeComment
			}
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
			modeCloser = str.SpacePrefix(rawLine) + "}"
		} else if currMode == modeError {
			// Inside Error mode
			line = newErrorLine(rawCount)
			// Close Error mode
			if str.TrimRightSpace(rawLine) == modeCloser {
				currMode = modeNone
			}
		} else {
			// Default: normal code line
			line = newCodeLine(rawCount)
			if strings.HasPrefix(rawLine, "func ") {
				codeType = classifyFunction(cleanLine)
				file.Block[CODE_FUNCTION] += 1
				file.Code[codeType] += 1
				line.CodeType = codeType
				currMode = modeFunction
				modeCloser = "}"
			} else if currMode == modeFunction {
				// Inside function mode
				line.CodeType = codeType
				// Close function mode
				if str.TrimRightSpace(rawLine) == modeCloser {
					currMode = modeNone
				}
			} else if strings.HasPrefix(rawLine, "type ") {
				codeType = classifyType(cleanLine)
				file.Block[CODE_TYPE] += 1
				file.Code[codeType] += 1
				line.CodeType = codeType
				if codeType == CODE_GROUP {
					file.Block[CODE_TYPE] -= 1 // undo increment
					currMode = modeTypeGroup
					modeCloser = ")"
				} else if codeType != PUB_ALIAS && codeType != PRIV_ALIAS {
					currMode = modeType
					modeCloser = "}"
				}
			} else if currMode == modeType {
				// Inside type mode
				line.CodeType = codeType
				// Close type mode
				if str.TrimRightSpace(rawLine) == modeCloser {
					currMode = modeNone
				}
			} else if currMode == modeTypeGroup {
				// Check if group closed
				if str.TrimRightSpace(rawLine) == modeCloser {
					line.CodeType = CODE_GROUP
					currMode = modeNone
					continue
				}
				// Inside type group mode
				codeType = classifyType("type " + cleanLine)
				file.Block[CODE_TYPE] += 1
				file.Code[codeType] += 1
				line.CodeType = codeType
			} else if strings.HasPrefix(rawLine, "const ") || strings.HasPrefix(rawLine, "var ") {
				codeType = classifyGlobal(cleanLine)
				file.Block[CODE_GLOBAL] += 1
				file.Code[codeType] += 1
				line.CodeType = codeType
				if codeType == CODE_GROUP {
					file.Block[CODE_GLOBAL] -= 1 // undo increment
					currMode = lang.Ternary(strings.HasPrefix(rawLine, "const "), modeConstGroup, modeVarGroup)
					modeCloser = ")"
				}
			} else if currMode == modeConstGroup || currMode == modeVarGroup {
				// Check if group closed
				if str.TrimRightSpace(rawLine) == modeCloser {
					line.CodeType = CODE_GROUP
					currMode = modeNone
					continue
				}
				// Inside global group mode
				codeType = classifyGlobal(lang.Ternary(currMode == modeConstGroup, "const ", "var ") + cleanLine)
				file.Block[CODE_GLOBAL] += 1
				file.Code[codeType] += 1
				line.CodeType = codeType
			} else {
				line.CodeType = NOT_CODE // unknown code type
			}
		}
		file.Lines = append(file.Lines, line)
	}

	return file, nil
}

// Create new Line with type: LINE_SPACE
func newSpaceLine() *Line {
	return &Line{Type: LINE_SPACE, Length: 1, CodeType: NOT_CODE}
}

// Create new Line with type: LINE_COMMENT
func newCommentLine(length int) *Line {
	return &Line{Type: LINE_COMMENT, Length: length, CodeType: NOT_CODE}
}

// Create new Line with type: LINE_HEAD
func newHeadLine(length int) *Line {
	return &Line{Type: LINE_HEAD, Length: length, CodeType: NOT_CODE}
}

// Create new Line with type: LINE_ERROR
func newErrorLine(length int) *Line {
	return &Line{Type: LINE_ERROR, Length: length, CodeType: NOT_CODE}
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
		mod.PackageGlobalBreakdown(),
		mod.PackageFunctionBreakdown(),
		mod.PackageTypeBreakdown(),
	}
	return strings.Join(out, "\n")
}

// Display CodeModule header
func (mod CodeModule) Header() string {
	blocks := mod.BreakdownBlocks()
	out := []string{
		fmt.Sprintf("Name: %s", mod.Name),
		fmt.Sprintf("Packages: %d", mod.CountValidNodes()),
		fmt.Sprintf("Files: %s", number.Comma(mod.CountFiles())),
		fmt.Sprintf("Lines: %s", number.Comma(mod.CountLines())),
		fmt.Sprintf("Chars: %s", number.Comma(mod.CountChars())),
		fmt.Sprintf("Globs: %s", number.Comma(blocks[CODE_GLOBAL])),
		fmt.Sprintf("Funcs: %s", number.Comma(blocks[CODE_FUNCTION])),
		fmt.Sprintf("Types: %s", number.Comma(blocks[CODE_TYPE])),
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
				percentDetails(usage[LINE_CODE], modUsage[LINE_CODE], pkgLineCount),
				percentDetails(usage[LINE_ERROR], modUsage[LINE_ERROR], pkgLineCount),
				percentDetails(usage[LINE_HEAD], modUsage[LINE_HEAD], pkgLineCount),
				percentDetails(usage[LINE_COMMENT], modUsage[LINE_COMMENT], pkgLineCount),
				percentDetails(usage[LINE_SPACE], modUsage[LINE_SPACE], pkgLineCount),
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
				percentDetails(usage[LINE_CODE], modUsage[LINE_CODE], pkgCharCount),
				percentDetails(usage[LINE_ERROR], modUsage[LINE_ERROR], pkgCharCount),
				percentDetails(usage[LINE_HEAD], modUsage[LINE_HEAD], pkgCharCount),
				percentDetails(usage[LINE_COMMENT], modUsage[LINE_COMMENT], pkgCharCount),
				percentDetails(usage[LINE_SPACE], modUsage[LINE_SPACE], pkgCharCount),
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

// Per package: breakdown function composition
func (mod CodeModule) PackageFunctionBreakdown() string {
	// Sort packages by total functions
	lookup := ds.NewLookupCode(mod.Packages)
	pkgNames := mod.PackageNames()
	funcCounts := list.Map(mod.Packages, func(pkg *Package) int {
		return pkg.CountBlocks(CODE_FUNCTION)
	})
	pkgFuncCounts := dict.Entries(dict.Zip(pkgNames, funcCounts))
	slices.SortFunc(pkgFuncCounts, sortDescCount)
	totalCount := list.Sum(funcCounts)

	out := []string{
		fmt.Sprintf("FUNCTIONS BREAKDOWN: %s", number.Comma(totalCount)),
		divider,
		fmt.Sprintf("%s|%s|%s|%s|%s|%s",
			str.Center("PubFunc", 10),
			str.Center("PrvFunc", 10),
			str.Center("PubMthod", 10),
			str.Center("PrvMthod", 10),
			str.Center("Total", 10),
			str.Center("Packages", 20),
		),
		divider,
	}

	modUsage := mod.BreakdownCodes()
	template := "%10s|%10s|%10s|%10s|%10s| %s"
	emptyPackages := make([]string, 0)
	for _, e := range pkgFuncCounts {
		pkgName, pkgFuncCount := e.Tuple()
		if pkgFuncCount == 0 {
			emptyPackages = append(emptyPackages, pkgName)
			continue // skip if empty
		}
		pkg := lookup[pkgName]
		usage := pkg.BreakdownCodes()
		out = append(out, fmt.Sprintf(template,
			number.Comma(usage[PUB_FUNCTION]),
			number.Comma(usage[PRIV_FUNCTION]),
			number.Comma(usage[PUB_METHOD]),
			number.Comma(usage[PRIV_METHOD]),
			number.Comma(pkgFuncCount),
			pkgName,
		))
		if mod.ShowDetails {
			out = append(out, fmt.Sprintf(template,
				percentDetails(usage[PUB_FUNCTION], modUsage[PUB_FUNCTION], pkgFuncCount),
				percentDetails(usage[PRIV_FUNCTION], modUsage[PRIV_FUNCTION], pkgFuncCount),
				percentDetails(usage[PUB_METHOD], modUsage[PUB_METHOD], pkgFuncCount),
				percentDetails(usage[PRIV_METHOD], modUsage[PRIV_METHOD], pkgFuncCount),
				percentage(pkgFuncCount, totalCount), "",
			))
		}
		out = append(out, divider)
	}

	out = append(out, fmt.Sprintf(template,
		number.Comma(modUsage[PUB_FUNCTION]),
		number.Comma(modUsage[PRIV_FUNCTION]),
		number.Comma(modUsage[PUB_METHOD]),
		number.Comma(modUsage[PRIV_METHOD]),
		number.Comma(totalCount),
		"TOTAL",
	))
	out = append(out, fmt.Sprintf(template,
		percentage(modUsage[PUB_FUNCTION], totalCount),
		percentage(modUsage[PRIV_FUNCTION], totalCount),
		percentage(modUsage[PUB_METHOD], totalCount),
		percentage(modUsage[PRIV_METHOD], totalCount),
		"", "",
	))
	out = append(out, divider)
	if len(emptyPackages) > 0 {
		out = append(out, fmt.Sprintf("Packages without functions: %d", len(emptyPackages)))
		for _, pkgName := range emptyPackages {
			out = append(out, "\t"+pkgName)
		}
		out = append(out, divider)
	}
	return strings.Join(out, "\n")
}

// Per package: breakdown global composition
func (mod CodeModule) PackageGlobalBreakdown() string {
	// Sort packages by total globals
	lookup := ds.NewLookupCode(mod.Packages)
	pkgNames := mod.PackageNames()
	globalCounts := list.Map(mod.Packages, func(pkg *Package) int {
		return pkg.CountBlocks(CODE_GLOBAL)
	})
	pkgGlobCounts := dict.Entries(dict.Zip(pkgNames, globalCounts))
	slices.SortFunc(pkgGlobCounts, sortDescCount)
	totalCount := list.Sum(globalCounts)

	out := []string{
		fmt.Sprintf("GLOBALS BREAKDOWN: %s", number.Comma(totalCount)),
		divider,
		fmt.Sprintf("%s|%s|%s|%s|%s|%s",
			str.Center("PubConst", 10),
			str.Center("PrvConst", 10),
			str.Center("PubVar", 10),
			str.Center("PrvVar", 10),
			str.Center("Total", 10),
			str.Center("Packages", 20),
		),
		divider,
	}

	modUsage := mod.BreakdownCodes()
	template := "%10s|%10s|%10s|%10s|%10s| %s"
	emptyPackages := make([]string, 0)
	for _, e := range pkgGlobCounts {
		pkgName, pkgGlobCount := e.Tuple()
		if pkgGlobCount == 0 {
			emptyPackages = append(emptyPackages, pkgName)
			continue // skip if empty
		}
		pkg := lookup[pkgName]
		usage := pkg.BreakdownCodes()
		out = append(out, fmt.Sprintf(template,
			number.Comma(usage[PUB_CONST]),
			number.Comma(usage[PRIV_CONST]),
			number.Comma(usage[PUB_VAR]),
			number.Comma(usage[PRIV_VAR]),
			number.Comma(pkgGlobCount),
			pkgName,
		))
		if mod.ShowDetails {
			out = append(out, fmt.Sprintf(template,
				percentDetails(usage[PUB_CONST], modUsage[PUB_CONST], pkgGlobCount),
				percentDetails(usage[PRIV_CONST], modUsage[PRIV_CONST], pkgGlobCount),
				percentDetails(usage[PUB_VAR], modUsage[PUB_VAR], pkgGlobCount),
				percentDetails(usage[PRIV_VAR], modUsage[PRIV_VAR], pkgGlobCount),
				percentage(pkgGlobCount, totalCount), "",
			))
		}
		out = append(out, divider)
	}

	out = append(out, fmt.Sprintf(template,
		number.Comma(modUsage[PUB_CONST]),
		number.Comma(modUsage[PRIV_CONST]),
		number.Comma(modUsage[PUB_VAR]),
		number.Comma(modUsage[PRIV_VAR]),
		number.Comma(totalCount),
		"TOTAL",
	))
	out = append(out, fmt.Sprintf(template,
		percentage(modUsage[PUB_CONST], totalCount),
		percentage(modUsage[PRIV_CONST], totalCount),
		percentage(modUsage[PUB_VAR], totalCount),
		percentage(modUsage[PRIV_VAR], totalCount),
		"", "",
	))
	out = append(out, divider)
	if len(emptyPackages) > 0 {
		out = append(out, fmt.Sprintf("Packages without globals: %d", len(emptyPackages)))
		for _, pkgName := range emptyPackages {
			out = append(out, "\t"+pkgName)
		}
		out = append(out, divider)
	}
	return strings.Join(out, "\n")
}

// Per package: breakdown function composition
func (mod CodeModule) PackageTypeBreakdown() string {
	// Sort packages by total types
	lookup := ds.NewLookupCode(mod.Packages)
	pkgNames := mod.PackageNames()
	typeCounts := list.Map(mod.Packages, func(pkg *Package) int {
		return pkg.CountBlocks(CODE_TYPE)
	})
	pkgTypeCounts := dict.Entries(dict.Zip(pkgNames, typeCounts))
	slices.SortFunc(pkgTypeCounts, sortDescCount)
	totalCount := list.Sum(typeCounts)

	out := []string{
		fmt.Sprintf("TYPES BREAKDOWN: %s", number.Comma(totalCount)),
		divider,
		fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s",
			str.Center("PubStrct", 8),
			str.Center("PrvStrct", 8),
			str.Center("PubIntfc", 8),
			str.Center("PrvIntfc", 8),
			str.Center("PubAlias", 8),
			str.Center("PrvAlias", 8),
			str.Center("Total", 10),
			str.Center("Packages", 20),
		),
		divider,
	}

	modUsage := mod.BreakdownCodes()
	template := "%8s|%8s|%8s|%8s|%8s|%8s|%10s| %s"
	emptyPackages := make([]string, 0)
	for _, e := range pkgTypeCounts {
		pkgName, pkgFuncCount := e.Tuple()
		if pkgFuncCount == 0 {
			emptyPackages = append(emptyPackages, pkgName)
			continue // skip if empty
		}
		pkg := lookup[pkgName]
		usage := pkg.BreakdownCodes()
		out = append(out, fmt.Sprintf(template,
			number.Comma(usage[PUB_STRUCT]),
			number.Comma(usage[PRIV_STRUCT]),
			number.Comma(usage[PUB_INTERFACE]),
			number.Comma(usage[PRIV_INTERFACE]),
			number.Comma(usage[PUB_ALIAS]),
			number.Comma(usage[PRIV_ALIAS]),
			number.Comma(pkgFuncCount),
			pkgName,
		))
		if mod.ShowDetails {
			out = append(out, fmt.Sprintf(template,
				percentDetails(usage[PUB_STRUCT], modUsage[PUB_STRUCT], pkgFuncCount),
				percentDetails(usage[PRIV_STRUCT], modUsage[PRIV_STRUCT], pkgFuncCount),
				percentDetails(usage[PUB_INTERFACE], modUsage[PUB_INTERFACE], pkgFuncCount),
				percentDetails(usage[PRIV_INTERFACE], modUsage[PRIV_INTERFACE], pkgFuncCount),
				percentDetails(usage[PUB_ALIAS], modUsage[PUB_ALIAS], pkgFuncCount),
				percentDetails(usage[PRIV_ALIAS], modUsage[PRIV_ALIAS], pkgFuncCount),
				percentage(pkgFuncCount, totalCount), "",
			))
		}
		out = append(out, divider)
	}

	out = append(out, fmt.Sprintf(template,
		number.Comma(modUsage[PUB_STRUCT]),
		number.Comma(modUsage[PRIV_STRUCT]),
		number.Comma(modUsage[PUB_INTERFACE]),
		number.Comma(modUsage[PRIV_INTERFACE]),
		number.Comma(modUsage[PUB_ALIAS]),
		number.Comma(modUsage[PRIV_ALIAS]),
		number.Comma(totalCount),
		"TOTAL",
	))
	out = append(out, fmt.Sprintf(template,
		percentage(modUsage[PUB_STRUCT], totalCount),
		percentage(modUsage[PRIV_STRUCT], totalCount),
		percentage(modUsage[PUB_INTERFACE], totalCount),
		percentage(modUsage[PRIV_INTERFACE], totalCount),
		percentage(modUsage[PUB_ALIAS], totalCount),
		percentage(modUsage[PRIV_ALIAS], totalCount),
		"", "",
	))
	out = append(out, divider)
	if len(emptyPackages) > 0 {
		out = append(out, fmt.Sprintf("Packages without types: %d", len(emptyPackages)))
		for _, pkgName := range emptyPackages {
			out = append(out, "\t"+pkgName)
		}
		out = append(out, divider)
	}
	return strings.Join(out, "\n")
}
