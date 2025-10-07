package main

import "github.com/roidaradal/fn/dict"

// Module info
type module struct {
	path         string
	name         string
	deps         []string
	tree         map[string]*fsnode
	independent  []string           // independent packages
	levels       map[int][]string   // non-independent package levels (0 = sink)
	externals    dict.StringListMap // external dependency: extPkg => list of packages that use it
	users        dict.StringListMap // internal dependency: intPkg => list of packages that use it
	dependencies dict.StringListMap // internal dependency: intPkg => list of packages it depends on
}

// File system folder
type fsnode struct {
	files   []string
	folders []string
}

// Create new module
func newModule() *module {
	return &module{
		path: "",
		name: "",
		deps: make([]string, 0),
		tree: make(map[string]*fsnode),
	}
}

// Create new fsnode
func newNode() *fsnode {
	return &fsnode{
		files:   make([]string, 0),
		folders: make([]string, 0),
	}
}
