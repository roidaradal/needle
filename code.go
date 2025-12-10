package main

import (
	"fmt"
	"maps"
	"path/filepath"
	"strings"

	"github.com/roidaradal/fn/conk"
	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/io"
	"github.com/roidaradal/fn/lang"
	"github.com/roidaradal/fn/str"
)

// Build the module tree, computing the dependencies, stats, and code composition
func buildModuleTree(mod *Module) error {
	type data struct {
		name string
		pkg  *Package
	}

	// Run concurrently
	task := func(entry NodeEntry) (data, error) {
		var d data
		name, node := entry.Tuple()
		pkg, err := newPackage(mod, name, node.Files)
		if err != nil {
			return d, err
		}
		return data{name, pkg}, nil
	}
	onReceive := func(d data) {
		mod.Packages = append(mod.Packages, d.pkg)
		mod.LineCount += d.pkg.LineCount
		mod.CharCount += d.pkg.CharCount
		dict.UpdateCounts(mod.Stats.Files, d.pkg.FileTypes)
		dict.UpdateCounts(mod.Stats.FileLines, d.pkg.FileLines)
		dict.UpdateCounts(mod.Stats.FileChars, d.pkg.FileChars)
		for dep, isInternal := range d.pkg.Deps {
			if isInternal {
				mod.Deps.Of[d.name] = append(mod.Deps.Of[d.name], dep)
			} else {
				mod.Deps.ExternalUsers[dep] = append(mod.Deps.ExternalUsers[dep], d.name)
			}
		}
	}
	entries := mod.packageNodeEntries()
	err := conk.Tasks(entries, task, onReceive)
	if err != nil {
		return err
	}

	// Compute inverse dependency => which package uses it
	mod.Deps.InternalUsers = dict.GroupByValueList(mod.Deps.Of)
	dict.SortValues(mod.Deps.InternalUsers)
	dict.SortValues(mod.Deps.ExternalUsers)
	dict.SortValues(mod.Deps.Of)

	// Update stats
	mod.Stats.Packages = dict.CounterFunc(mod.Packages, func(pkg *Package) PackageType {
		return pkg.Type
	})
	return nil
}

// Build Package object for given name
func newPackage(mod *Module, name string, files []string) (*Package, error) {
	folder := mod.Path + name
	pkg := &Package{
		Name:      str.GuardWith(strings.TrimPrefix(name, "/"), "/"),
		Files:     make([]*File, 0),
		Deps:      make(map[string]bool),
		Blocks:    make(dict.Counter[BlockType]),
		Codes:     make(dict.Counter[CodeType]),
		FileTypes: make(dict.Counter[FileType]),
		FileLines: make(dict.Counter[FileType]),
		FileChars: make(dict.Counter[FileType]),
	}

	// Run concurrently
	task := func(filename string) (*File, error) {
		path := filepath.Join(folder, filename)
		return newFile(mod, pkg, path)
	}
	onReceive := func(file *File) {
		pkg.Files = append(pkg.Files, file)
		maps.Copy(pkg.Deps, file.Deps)
		dict.UpdateCounts(pkg.Blocks, file.Blocks)
		dict.UpdateCounts(pkg.Codes, file.Codes)
	}
	err := conk.Tasks(files, task, onReceive)
	if err != nil {
		return nil, err
	}

	// Set file type stats
	pkg.FileTypes = dict.CounterFunc(pkg.Files, func(f *File) FileType {
		return f.Type
	})
	// Set file line stats
	for _, f := range pkg.Files {
		numLines := len(f.Lines)
		pkg.FileLines[f.Type] += numLines
		pkg.LineCount += numLines
		numChars := f.CharCount
		pkg.FileChars[f.Type] += numChars
		pkg.CharCount += numChars
	}
	return pkg, nil
}

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

// Build File object for given file path
func newFile(mod *Module, pkg *Package, path string) (*File, error) {
	if !io.PathExists(path) {
		return nil, fmt.Errorf("file %q does not exist", path)
	}

	file := &File{
		Name:   filepath.Base(path),
		Type:   lang.Ternary(endsWith(path, "_test.go"), FILE_TEST, FILE_CODE),
		Lines:  make([]*Line, 0),
		Deps:   make(map[string]bool),
		Blocks: make(dict.Counter[BlockType]),
		Codes:  make(dict.Counter[CodeType]),
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
		} else if startsWith(cleanLine, "//") {
			// Single-Line Comment
			line = newCommentLine(rawCount)
		} else if startsWith(cleanLine, "/*") {
			// Start Multi-Line Comment
			line = newCommentLine(rawCount)
			// Set to comment mode if not single-line block comment /*..*/
			if !endsWith(cleanLine, "*/") {
				currMode = modeComment
				modeCloser = "*/"
			}
		} else if currMode == modeComment {
			// Inside Multi-Line comment
			line = newCommentLine(rawCount)
			// Close Multi-Line comment
			if endsWith(cleanLine, modeCloser) {
				currMode = modeNone
			}
		} else if startsWith(cleanLine, "package ") {
			// Package header
			line = newHeadLine(rawCount)
			if name, ok := getLinePart(cleanLine, 1); ok && pkg.Type == "" {
				pkg.Type = lang.Ternary(name == "main", PKG_MAIN, PKG_LIB)
			}
		} else if startsWith(cleanLine, "import ") {
			// Import header
			line = newHeadLine(rawCount)
			if endsWith(cleanLine, " (") {
				currMode = modeHead
				modeCloser = ")"
			} else {
				// Single line import, get last part
				// to consider import aliases (e.g. import a2 "pkg/a")
				if dep, ok := getLinePart(cleanLine, -1); ok {
					file.addDependency(mod, dep)
				}
			}
		} else if currMode == modeHead {
			// Inside Header mode
			line = newHeadLine(rawCount)
			if cleanLine == modeCloser {
				currMode = modeNone
			} else {
				if dep, ok := getLinePart(cleanLine, -1); ok {
					file.addDependency(mod, dep)
				}
			}
		} else if cleanLine == errLine || endsWith(cleanLine, errSuffix) {
			// Start error mode
			line = newErrorLine(rawCount)
			currMode = modeError
			modeCloser = str.SpacePrefix(rawLine) + "}"
		} else if currMode == modeError {
			// Inside error mode
			line = newErrorLine(rawCount)
			// Close error mode
			if str.TrimRightSpace(rawLine) == modeCloser {
				currMode = modeNone
			}
		} else {
			// Default: normal code line
			line = newCodeLine(rawCount)
			if startsWith(rawLine, "func ") {
				// Start function
				codeType = classifyFunction(cleanLine)
				line.CodeType = codeType
				file.Blocks[CODE_FUNCTION] += 1
				file.Codes[codeType] += 1
				currMode = modeFunction
				modeCloser = "}"
			} else if currMode == modeFunction {
				// Inside function mode
				line.CodeType = codeType
				// Close function mode
				if str.TrimRightSpace(rawLine) == modeCloser {
					currMode = modeNone
				}
			} else if startsWith(rawLine, "type ") {
				// Start type
				codeType = classifyType(cleanLine)
				line.CodeType = codeType
				if codeType == CODE_GROUP {
					currMode = modeTypeGroup
					modeCloser = ")"
				} else {
					file.Blocks[CODE_TYPE] += 1
					file.Codes[codeType] += 1
					if codeType != PUB_ALIAS && codeType != PRIV_ALIAS {
						currMode = modeType
						modeCloser = "}"
					}
				}
			} else if currMode == modeType {
				// Inside type mode
				line.CodeType = codeType
				// Close type mode
				if str.TrimRightSpace(rawLine) == modeCloser {
					currMode = modeNone
				}
			} else if currMode == modeTypeGroup {
				// Check if type group closed
				if str.TrimRightSpace(rawLine) == modeCloser {
					line.CodeType = CODE_GROUP
					currMode = modeNone
				} else {
					// Inside type group mode
					codeType = classifyType("type " + cleanLine)
					line.CodeType = codeType
					file.Blocks[CODE_TYPE] += 1
					file.Codes[codeType] += 1
				}
			} else if startsWith(rawLine, "const ") || startsWith(rawLine, "var ") {
				// Constant or variable
				codeType = classifyGlobal(cleanLine)
				line.CodeType = codeType
				if codeType == CODE_GROUP {
					currMode = lang.Ternary(startsWith(rawLine, "const "), modeConstGroup, modeVarGroup)
					modeCloser = ")"
				} else {
					file.Blocks[CODE_GLOBAL] += 1
					file.Codes[codeType] += 1
				}
			} else if currMode == modeConstGroup || currMode == modeVarGroup {
				// Check if const/var group closed
				if str.TrimRightSpace(rawLine) == modeCloser {
					line.CodeType = CODE_GROUP
					currMode = modeNone
				} else {
					// Inside global group mode
					prefix := lang.Ternary(currMode == modeConstGroup, "const ", "var ")
					codeType = classifyGlobal(prefix + cleanLine)
					line.CodeType = codeType
					file.Blocks[CODE_GLOBAL] += 1
					file.Codes[codeType] += 1
				}
			} else {
				line.CodeType = NOT_CODE // unknown code type
			}
		}

		file.Lines = append(file.Lines, line)
		file.CharCount += line.Length
	}

	return file, nil
}

// Classify function as (public/private) x (function/method)
func classifyFunction(line string) CodeType {
	parts := str.SpaceSplit(line)
	isMethod := parts[1][0] == '(' // check if has function receiver
	if isMethod {
		parts = str.CleanSplit(line, ")")
		isPublic := str.StartsWithUpper(parts[1])
		return lang.Ternary(isPublic, PUB_METHOD, PRIV_METHOD)
	} else {
		isPublic := str.StartsWithUpper(parts[1])
		return lang.Ternary(isPublic, PUB_FUNCTION, PRIV_FUNCTION)
	}
}

// Classify type as (public/private) x (struct/interface/alias) or code block
func classifyType(line string) CodeType {
	parts := str.SpaceSplit(line)
	if parts[1] == "(" {
		return CODE_GROUP
	}
	isPublic := str.StartsWithUpper(parts[1])
	if endsWith(line, " interface {") {
		return lang.Ternary(isPublic, PUB_INTERFACE, PRIV_INTERFACE)
	} else if endsWith(line, " struct {") {
		return lang.Ternary(isPublic, PUB_STRUCT, PRIV_STRUCT)
	} else {
		return lang.Ternary(isPublic, PUB_ALIAS, PRIV_ALIAS)
	}
}

// Classify globals as (public/private) x (const/var) or code block
func classifyGlobal(line string) CodeType {
	parts := str.SpaceSplit(line)
	if parts[1] == "(" {
		return CODE_GROUP
	}
	isPublic := str.StartsWithUpper(parts[1])
	if parts[0] == "const" {
		return lang.Ternary(isPublic, PUB_CONST, PRIV_CONST)
	} else {
		return lang.Ternary(isPublic, PUB_VAR, PRIV_VAR)
	}
}
