package needle

import (
	"fmt"

	"github.com/roidaradal/fn/dict"
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

// Node string representation
func (n Node) String() string {
	return fmt.Sprintf("<Files: %d, Folders: %d>", len(n.Files), len(n.Folders))
}

// Go module
type Module struct {
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
