package needle

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/ds"
	"github.com/roidaradal/fn/io"
	"github.com/roidaradal/fn/list"
)

// Create Module object for Go module at path
func baseModule(cfg *Config) (*Module, error) {
	mod, err := newModuleAt(cfg.Path)
	if err != nil {
		return nil, err
	}
	mod.IsCompact = cfg.IsCompact

	decorators := []func(*Module) error{
		readGoModFile,
		buildModuleTree,
	}
	err = applyDecorators(mod, decorators)
	if err != nil {
		return nil, err
	}

	return mod, nil
}

// Check if path is a valid directory, create Module object for path
func newModuleAt(path string) (*Module, error) {
	// Remove trailing slash if necessary
	path = strings.TrimSuffix(path, "/")
	// Check if path is directory
	if !io.IsDir(path) {
		return nil, fmt.Errorf("path '%s' is not a directory", path)
	}
	mod := newModule()
	mod.Path = path
	return mod, nil
}

// Read go.mod file and get module name and external dependencies
func readGoModFile(mod *Module) error {
	path := fmt.Sprintf("%s/go.mod", mod.Path)
	if !io.PathExists(path) {
		return fmt.Errorf("file '%s' does not exist", path)
	}

	lines, err := io.ReadNonEmptyLines(path)
	if err != nil {
		return err
	}

	depMode := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			// Get module name
			if name, ok := getLinePart(line, 1); ok {
				mod.Name = name
			}
		} else if strings.HasPrefix(line, "require ") {
			if strings.HasSuffix(line, "(") {
				// toggle depMode on
				depMode = true
			} else if isDirectDependency(line) {
				// direct dependency = add to external dependencies
				if extPkg, ok := getLinePart(line, 1); ok {
					mod.ExternalDeps = append(mod.ExternalDeps, extPkg)
				}
			}
		} else if line == ")" {
			// toggle depMode off
			depMode = false
		} else if depMode && isDirectDependency(line) {
			// depMode and direct dependency = add to external dependencies
			if extPkg, ok := getLinePart(line, 0); ok {
				mod.ExternalDeps = append(mod.ExternalDeps, extPkg)
			}
		}
	}
	return nil
}

// Build module tree, going through folders and subfolders
func buildModuleTree(mod *Module) error {
	rootNode, err := buildNode(mod.Path)
	if err != nil {
		return err
	}
	mod.Tree["/"] = rootNode

	rootFolders := list.Map(rootNode.Folders, func(folder string) string {
		return fmt.Sprintf("/%s", folder)
	})
	q := ds.QueueFrom(rootFolders)
	for q.NotEmpty() {
		folder, _ := q.Dequeue()
		path := mod.Path + folder
		node, err := buildNode(path)
		if err != nil {
			return err
		}
		if len(node.Files) > 0 {
			mod.Tree[folder] = node
		}
		for _, subFolder := range node.Folders {
			q.Enqueue(fmt.Sprintf("%s/%s", folder, subFolder))
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
		} else if strings.HasSuffix(name, ".go") {
			node.Files = append(node.Files, name)
		}
	}
	return node, nil
}

// Module string representation
func (mod Module) String() string {
	out := []string{
		fmt.Sprintf("Name: %s", mod.Name),
	}
	out = append(out, fmt.Sprintf("Tree: %d / %d", mod.CountValidNodes(), len(mod.Tree)))
	if !mod.IsCompact {
		keys := dict.Keys(mod.Tree)
		slices.Sort(keys)
		maxLength := getMaxLength(keys)
		template := fmt.Sprintf("\t%%-%ds : %%s", maxLength)
		for _, key := range keys {
			line := fmt.Sprintf(template, key, mod.Tree[key])
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

// Count valid Nodes = len(Files) > 0
func (mod Module) CountValidNodes() int {
	validNodes := list.Filter(dict.Values(mod.Tree), func(node *Node) bool {
		return len(node.Files) > 0
	})
	return len(validNodes)
}

// Valid Tree Entries = len(node.Files) > 0
func (mod Module) ValidTreeEntries() []TreeEntry {
	entries := dict.Entries(mod.Tree)
	entries = list.Filter(entries, func(e TreeEntry) bool {
		node := e.Value
		return len(node.Files) > 0
	})
	return entries
}
