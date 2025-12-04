package main

import "github.com/roidaradal/fn/dict"

// Go module
type Module struct {
	Path  string           // Go module filesystem path
	Name  string           // Go module name
	Nodes map[string]*Node // Mapping of subfolders to Node inside Go module
	Deps
	Stats
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

// Create new Module
func newModule() *Module {
	return &Module{
		Nodes: make(map[string]*Node),
		Deps: Deps{
			Of:            make(dict.StringListMap),
			InternalUsers: make(dict.StringListMap),
			ExternalUsers: make(dict.StringListMap),
			Independent:   make([]string, 0),
			Levels:        make(map[int][]string),
		},
		Stats: Stats{},
	}
}

type (
	PackageType string
	FileType    string
	LineType    string
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
	Name  string
	Type  PackageType
	Files []*File
}

// Go File object
type File struct {
	Name  string
	Type  FileType
	Lines []*Line
}

// Go Line object
type Line struct {
	Type   LineType
	Length int // character count
}

// Package code for lookup
func (pkg Package) GetCode() string {
	return pkg.Name
}

// File code for lookup
func (f File) GetCode() string {
	return f.Name
}
