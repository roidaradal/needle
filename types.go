package main

import (
	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/list"
)

// Go module
type Module struct {
	Path     string           // Go module filesystem path
	Name     string           // Go module name
	Nodes    map[string]*Node // Mapping of subfolders to Node inside Go module
	Packages []*Package       // list of Package objects
	Deps
	Stats
	Code
}

// Dependencies info
type Deps struct {
	Of            dict.StringListMap // Internal package => list of subpackges it depends on
	InternalUsers dict.StringListMap // Internal package => list of subpackages that directly use it
	ExternalUsers dict.StringListMap // External dependency => list of subpackages that directly use it
	External      []string           // List of external subpackages
	Independent   []string           // List of independent subpackages (not in dependency DAG)
	Levels        map[int][]string   // Non-independent subpackage levels (0 = sink)
	Nodes         dict.StringMap     // Non-independent subpackage => {x: xPosition, y: yPosition}
	Edges         []string           // List of node1-node2 edges
}

// Stats info
type Stats struct {
	PackageCount int
	FileCount    int
	LineCount    int
	CharCount    int
	Packages     dict.Counter[PackageType]
	Files        dict.Counter[FileType]
	FileLines    dict.Counter[FileType]
	FileChars    dict.Counter[FileType]
}

// Code info
type Code struct {
	Blocks dict.Counter[BlockType]
	Types  dict.Counter[CodeType]
	Lines  dict.Counter[LineType]
	Chars  dict.Counter[LineType]
}

// Create new Module
func newModule() *Module {
	return &Module{
		Nodes:    make(map[string]*Node),
		Packages: make([]*Package, 0),
		Deps: Deps{
			Of:            make(dict.StringListMap),
			InternalUsers: make(dict.StringListMap),
			ExternalUsers: make(dict.StringListMap),
			External:      make([]string, 0),
			Independent:   make([]string, 0),
			Levels:        make(map[int][]string),
		},
		Code: Code{
			Blocks: make(dict.Counter[BlockType]),
			Types:  make(dict.Counter[CodeType]),
			Lines:  make(dict.Counter[LineType]),
			Chars:  make(dict.Counter[LineType]),
		},
		Stats: Stats{
			Packages:  make(dict.Counter[PackageType]),
			Files:     make(dict.Counter[FileType]),
			FileLines: make(dict.Counter[FileType]),
			FileChars: make(dict.Counter[FileType]),
		},
	}
}

type (
	NodeEntry  = dict.Entry[string, *Node]
	CountEntry = dict.Entry[string, int]
)

type (
	PackageType string
	FileType    string
	LineType    string
	BlockType   string
	CodeType    string
)

const (
	PKG_LIB  PackageType = "Lib"
	PKG_MAIN PackageType = "Main"
)
const (
	FILE_CODE FileType = "Code"
	FILE_TEST FileType = "Test"
)
const (
	LINE_CODE    LineType = "Codes"
	LINE_ERROR   LineType = "Error"
	LINE_HEAD    LineType = "Head"
	LINE_COMMENT LineType = "Comment"
	LINE_SPACE   LineType = "Space"
)

const (
	CODE_FUNCTION BlockType = "Function"
	CODE_TYPE     BlockType = "Type"
	CODE_GLOBAL   BlockType = "Global"
)

const (
	NOT_CODE       CodeType = "not_code"
	CODE_GROUP     CodeType = "group"
	PUB_FUNCTION   CodeType = "PubFunction"
	PRIV_FUNCTION  CodeType = "PrivFunction"
	PUB_METHOD     CodeType = "PubMethod"
	PRIV_METHOD    CodeType = "PrivMethod"
	PUB_STRUCT     CodeType = "PubStruct"
	PRIV_STRUCT    CodeType = "PrivStruct"
	PUB_INTERFACE  CodeType = "PubInterface"
	PRIV_INTERFACE CodeType = "PrivInterface"
	PUB_ALIAS      CodeType = "PubAlias"
	PRIV_ALIAS     CodeType = "PrivAlias"
	PUB_CONST      CodeType = "PubConst"
	PRIV_CONST     CodeType = "PrivConst"
	PUB_VAR        CodeType = "PubVar"
	PRIV_VAR       CodeType = "PrivVar"
)

// File system folder
type Node struct {
	Folders []string
	Files   []string
}

// Create new Node
func newNode() *Node {
	return &Node{
		Folders: make([]string, 0),
		Files:   make([]string, 0),
	}
}

// Get number of files for Node
func (n Node) FileCount() int {
	return len(n.Files)
}

// Go Package object
type Package struct {
	Name      string
	Type      PackageType
	Files     []*File
	Deps      map[string]bool // dependency => isInternal
	Blocks    dict.Counter[BlockType]
	Codes     dict.Counter[CodeType]
	FileTypes dict.Counter[FileType]
	FileLines dict.Counter[FileType]
	FileChars dict.Counter[FileType]
	LineTypes dict.Counter[LineType]
	CharTypes dict.Counter[LineType]
	LineCount int
	CharCount int
}

// Go File object
type File struct {
	Name      string
	Type      FileType
	Lines     []*Line
	Deps      map[string]bool // dependency => isInternal
	Blocks    dict.Counter[BlockType]
	Codes     dict.Counter[CodeType]
	LineTypes dict.Counter[LineType]
	CharTypes dict.Counter[LineType]
	CharCount int
}

// Go Line object
type Line struct {
	Type   LineType
	Length int // character count
	CodeType
}

// Package code for lookup
func (pkg Package) GetCode() string {
	return pkg.Name
}

// File code for lookup
func (f File) GetCode() string {
	return f.Name
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

// Return module package names
func (mod Module) PackageNames() []string {
	return list.Map(mod.Packages, (*Package).GetCode)
}

// Return package file names
func (pkg Package) FileNames() []string {
	return list.Map(pkg.Files, (*File).GetCode)
}

// Return package file count
func (pkg *Package) FileCount() int {
	return len(pkg.Files)
}
