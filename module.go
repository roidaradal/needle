package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/ds"
	"github.com/roidaradal/fn/io"
	"github.com/roidaradal/fn/list"
)

// Build Module object for Go module at path
func BuildModule(path string) (*Module, error) {
	// Remove trailing slash if necessary
	path = strings.TrimSuffix(path, "/")
	// Check if path is directory
	if !io.IsDir(path) {
		return nil, fmt.Errorf("path %q is not a directory", path)
	}

	// Create Module object
	mod := newModule()
	mod.Path = path

	// Apply decorator functions to Module
	decorators := []func(*Module) error{
		readGoModFile,           // module
		buildModuleNodes,        // module
		buildModuleTree,         // module
		computeDependencyLevels, // deps
		computeDependencyLayout, // deps
	}
	for _, decorator := range decorators {
		err := decorator(mod)
		if err != nil {
			return nil, err
		}
	}

	return mod, nil
}

// Read go.mod file to get module name and list of direct, external dependencies
func readGoModFile(mod *Module) error {
	// Ensure go.mod file exists
	path := filepath.Join(mod.Path, "go.mod")
	if !io.PathExists(path) {
		return fmt.Errorf("file %q does not exist", path)
	}

	// Read go.mod lines
	lines, err := io.ReadNonEmptyLines(path)
	if err != nil {
		return err
	}

	inDepMode := false // true if inside dependency block
	for _, line := range lines {
		if startsWith(line, "module ") {
			// Get module name
			if name, ok := getLinePart(line, 1); ok {
				mod.Name = name
			}
		} else if startsWith(line, "require ") {
			if endsWith(line, "(") {
				inDepMode = true // toggle depMode on
			} else if isDirectDependency(line) {
				// add direct, external dependency
				if extPkg, ok := getLinePart(line, 1); ok {
					mod.addExternalDependency(extPkg)
				}
			}
		} else if line == ")" {
			inDepMode = false // toggle depMode off
		} else if inDepMode && isDirectDependency(line) {
			// DepMode: add direct, external dependency
			if extPkg, ok := getLinePart(line, 0); ok {
				mod.addExternalDependency(extPkg)
			}
		}
	}
	return nil
}

// Build module nodes, going through folders and subfolders
func buildModuleNodes(mod *Module) error {
	rootNode, err := buildNode(mod.Path)
	if err != nil {
		return err
	}
	mod.Nodes["/"] = rootNode
	rootFileCount := rootNode.FileCount()
	if rootFileCount > 0 {
		mod.Stats.PackageCount += 1
		mod.Stats.FileCount += rootFileCount
	}

	q := ds.QueueFrom(list.Map(rootNode.Folders, func(folder string) string {
		return joinPath("", folder)
	}))
	for q.NotEmpty() {
		folder, _ := q.Dequeue()
		node, err := buildNode(mod.Path + folder)
		if err != nil {
			return err
		}
		nodeFileCount := node.FileCount()
		if nodeFileCount > 0 {
			mod.Nodes[folder] = node
			mod.Stats.PackageCount += 1
			mod.Stats.FileCount += nodeFileCount
		}
		for _, subFolder := range node.Folders {
			q.Enqueue(joinPath(folder, subFolder))
		}
	}
	return nil
}

// Build Node for given folder path, get subfolders and .go files
func buildNode(path string) (*Node, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	node := newNode()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() && isPublicFolder(name) {
			node.Folders = append(node.Folders, name)
		} else if endsWith(name, ".go") {
			node.Files = append(node.Files, name)
		}
	}
	return node, nil
}

// Package node entries: nodes with at least 1 file
func (mod Module) packageNodeEntries() []NodeEntry {
	entries := dict.Entries(mod.Nodes)
	entries = list.Filter(entries, func(e NodeEntry) bool {
		node := e.Value
		return len(node.Files) > 0
	})
	return entries
}

// Add external package dependency
func (mod *Module) addExternalDependency(extPkg string) {
	mod.Deps.External = append(mod.Deps.External, extPkg)
	mod.Deps.ExternalUsers[extPkg] = make([]string, 0)
}

// Check if dependency line does not end with // indirect
func isDirectDependency(line string) bool {
	return !endsWith(line, "// indirect")
}

// Check if directory path doesn't start with dot, underscore, dash
func isPublicFolder(name string) bool {
	prefixes := []string{".", "_", "-"}
	return list.All(prefixes, func(prefix string) bool {
		return !startsWith(name, prefix)
	})
}
