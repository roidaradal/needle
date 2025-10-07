package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/roidaradal/fn"
	"github.com/roidaradal/fn/ds"
	"github.com/roidaradal/fn/io"
	"github.com/roidaradal/fn/str"
)

const indirect string = "// indirect"

type moduleDecorator = func(*module) error

// Extract information from the given module path
func getModuleInfo(folder string) error {
	mod, err := validateModuleFolder(folder)
	if err != nil {
		return err
	}

	decorators := []moduleDecorator{
		readGoModFile,
		getModuleContents,
		buildModuleTree,
	}

	for _, decorator := range decorators {
		err = decorator(mod)
		if err != nil {
			return err
		}
	}

	fmt.Println(mod)
	return nil
}

// Check if folder name does not start with dot, underscore, dash
func isPublicFolder(folder string) bool {
	ok1 := !strings.HasPrefix(folder, ".")
	ok2 := !strings.HasPrefix(folder, "_")
	ok3 := !strings.HasPrefix(folder, "-")
	return ok1 && ok2 && ok3
}

// Transform the module path and check if valid directory
func validateModuleFolder(folder string) (*module, error) {
	// Remove trailing slash if necessary
	folder = strings.TrimSuffix(folder, "/")

	// TODO: replace with io.IsDir(path)
	// Check if path is directory
	f, err := os.Stat(folder)
	if err != nil {
		return nil, err
	}
	if !f.IsDir() {
		return nil, fmt.Errorf("path '%s' is not a directory", folder)
	}

	mod := newModule()
	mod.path = folder
	return mod, nil
}

// Read go.mod file and get module name and dependencies
func readGoModFile(mod *module) error {
	path := fmt.Sprintf("%s/go.mod", mod.path)
	if !io.PathExists(path) {
		return fmt.Errorf("file '%s' does not exist", path)
	}

	lines, err := io.ReadLines(path)
	if err != nil {
		return err
	}

	var parts []string
	depMode := false
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			// Get module name
			parts = str.SpaceSplit(line)
			if len(parts) >= 2 {
				mod.name = parts[1]
			}
		} else if strings.HasPrefix(line, "require ") {
			if strings.HasSuffix(line, "(") {
				depMode = true // toggle dep mode on
			} else if !strings.HasSuffix(line, indirect) {
				// if direct requirement, add to dependency list
				parts = str.SpaceSplit(line)
				if len(parts) >= 2 {
					mod.deps = append(mod.deps, parts[1])
				}
			}
		} else if line == ")" {
			depMode = false // toggle dep mode off
		} else if depMode {
			// dep mode and direct requirement = add to dependency list
			if !strings.HasSuffix(line, indirect) {
				parts = str.SpaceSplit(line)
				if len(parts) >= 1 {
					mod.deps = append(mod.deps, parts[0])
				}
			}
		}
	}
	return nil
}

// Get list of go files and top-level folders in module folder
func getModuleContents(mod *module) error {
	rootNode, err := buildNode(mod.path)
	if err != nil {
		return err
	}
	mod.tree["/"] = rootNode
	return nil
}

// Build the module tree, going through folders and subfolders
func buildModuleTree(mod *module) error {
	folders := fn.Map(mod.tree["/"].folders, func(folder string) string {
		return fmt.Sprintf("/%s/", folder)
	})
	q := ds.QueueFrom(folders)
	for !q.IsEmpty() {
		folder, _ := q.Dequeue()
		path := mod.path + folder
		n, err := buildNode(path)
		if err != nil {
			return err
		}
		mod.tree[folder] = n
		for _, subFolder := range n.folders {
			q.Enqueue(fmt.Sprintf("%s%s/", folder, subFolder))
		}
	}
	return nil
}

// Build the node for the given folder
func buildNode(folder string) (*node, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, err
	}
	n := newNode()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() && isPublicFolder(name) {
			n.folders = append(n.folders, name)
		} else if strings.HasSuffix(name, ".go") {
			n.files = append(n.files, name)
		}
	}
	return n, nil
}
