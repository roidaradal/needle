package main

import "github.com/roidaradal/fn/dict"

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
	Independent   []string           // List of independent subpackages (not in dependency DAG)
	Levels        map[int][]string   // Non-independent subpackage levels (0 = sink)
}

// Stats info
type Stats struct {
	PackageCount int
	FileCount    int
}

// Code info
type Code struct {
	Blocks map[BlockType]int
	Types  map[CodeType]int
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
			Independent:   make([]string, 0),
			Levels:        make(map[int][]string),
		},
		Code: Code{
			Blocks: make(map[BlockType]int),
			Types:  make(map[CodeType]int),
		},
		Stats: Stats{},
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
	PKG_LIB  PackageType = "lib"
	PKG_MAIN PackageType = "main"
)
const (
	FILE_CODE FileType = "code"
	FILE_TEST FileType = "test"
)
const (
	LINE_CODE    LineType = "code"
	LINE_ERROR   LineType = "error"
	LINE_HEAD    LineType = "head"
	LINE_SPACE   LineType = "space"
	LINE_COMMENT LineType = "comment"
)

const (
	CODE_FUNCTION BlockType = "function"
	CODE_TYPE     BlockType = "type"
	CODE_GLOBAL   BlockType = "global"
)

const (
	NOT_CODE       CodeType = "not_code"
	CODE_GROUP     CodeType = "group"
	PUB_FUNCTION   CodeType = "pub_function"
	PRIV_FUNCTION  CodeType = "priv_function"
	PUB_METHOD     CodeType = "pub_method"
	PRIV_METHOD    CodeType = "priv_method"
	PUB_STRUCT     CodeType = "pub_struct"
	PRIV_STRUCT    CodeType = "priv_struct"
	PUB_INTERFACE  CodeType = "pub_interface"
	PRIV_INTERFACE CodeType = "priv_interface"
	PUB_ALIAS      CodeType = "pub_alias"
	PRIV_ALIAS     CodeType = "priv_alias"
	PUB_CONST      CodeType = "pub_const"
	PRIV_CONST     CodeType = "priv_const"
	PUB_VAR        CodeType = "pub_var"
	PRIV_VAR       CodeType = "priv_var"
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
	Name   string
	Type   PackageType
	Files  []*File
	Deps   map[string]bool // dependency => isInternal
	Blocks map[BlockType]int
	Codes  map[CodeType]int
}

// Go File object
type File struct {
	Name   string
	Type   FileType
	Lines  []*Line
	Deps   map[string]bool // dependency => isInternal
	Blocks map[BlockType]int
	Codes  map[CodeType]int
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
