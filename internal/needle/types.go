package needle

import (
	"fmt"

	"github.com/roidaradal/fn/dict"
)

// Config object
type Config struct {
	Option      string
	Path        string
	ShowDetails bool
}

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

// Node string representation
func (n Node) String() string {
	return fmt.Sprintf("<Files: %d, Folders: %d>", len(n.Files), len(n.Folders))
}

// Go module
type Module struct {
	ShowDetails  bool
	Path         string           // Go module path in filesystem
	Name         string           // Go module name
	Tree         map[string]*Node // Mapping of subfolders to Node inside Go module
	ExternalDeps []string         // List of external package dependencies used by Go module
}

// Create new Module
func newModule() *Module {
	return &Module{
		Tree:         make(map[string]*Node),
		ExternalDeps: make([]string, 0),
	}
}

// Dependencies Module
type DepsModule struct {
	*Module
	ExternalUsers    dict.StringListMap // External dependency: externalPkg => list of subpackages that use it
	InternalUsers    dict.StringListMap // Internal dependency: internalPkg => list of subpackages that use it
	DependenciesOf   dict.StringListMap // Internal dependency: internalPkg => list of subpackages it depends on
	IndependentSubs  []string           // List of independent subpackages (not part of the dependency DAG)
	DependencyLevels map[int][]string   // Non-independent subpackage levels (0 = sink)
}

// Create new DepsModule
func newDepsModule(mod *Module) *DepsModule {
	if mod == nil {
		mod = newModule()
	}
	return &DepsModule{
		Module:           mod,
		ExternalUsers:    make(dict.StringListMap),
		InternalUsers:    make(dict.StringListMap),
		DependenciesOf:   make(dict.StringListMap),
		IndependentSubs:  make([]string, 0),
		DependencyLevels: make(map[int][]string),
	}
}

// Stats Module
type StatsModule struct {
	*Module
	Packages []*Package
}

// Create new StatsModule
func newStatsModule(mod *Module) *StatsModule {
	if mod == nil {
		mod = newModule()
	}
	return &StatsModule{
		Module:   mod,
		Packages: make([]*Package, 0),
	}
}

// Code Module
type CodeModule struct {
	*Module
	Packages []*Package
}

// Create new CodeModule
func newCodeModule(mod *Module) *CodeModule {
	if mod == nil {
		mod = newModule()
	}
	return &CodeModule{
		Module:   mod,
		Packages: make([]*Package, 0),
	}
}

type (
	PackageType string
	FileType    string
	LineType    string
	CodeType    string
	BlockType   string
)

const (
	PKG_MAIN     PackageType = "main"
	PKG_LIB      PackageType = "lib"
	FILE_CODE    FileType    = "code"
	FILE_TEST    FileType    = "test"
	LINE_CODE    LineType    = "code"
	LINE_ERROR   LineType    = "error"
	LINE_HEAD    LineType    = "head"
	LINE_SPACE   LineType    = "space"
	LINE_COMMENT LineType    = "comment"
)

// Package object
type Package struct {
	Name  string
	Type  PackageType
	Files []*File
}

// File object
type File struct {
	Name  string
	Type  FileType
	Lines []*Line
	Block map[BlockType]int
	Code  map[CodeType]int
}

// Line object
type Line struct {
	CodeType
	Type   LineType
	Length int // numChars
}

// Package code for lookup
func (pkg Package) GetCode() string {
	return pkg.Name
}

// File code for lookup
func (f File) GetCode() string {
	return f.Name
}

type (
	TreeEntry  = dict.Entry[string, *Node]
	CountEntry = dict.Entry[string, int]
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

const (
	CODE_FUNCTION BlockType = "function"
	CODE_TYPE     BlockType = "type"
	CODE_GLOBAL   BlockType = "global"
)
