package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/roidaradal/fn/dict"
)

const indent string = "  " // two spaces

type module struct {
	path string
	name string
	deps []string
	tree map[string]*node
}

type node struct {
	files   []string
	folders []string
}

func (m module) String() string {
	out := make([]string, 0)
	out = append(out, fmt.Sprintf("Path: %s", m.path))
	out = append(out, fmt.Sprintf("Name: %s", m.name))
	out = append(out, fmt.Sprintf("Deps: %d", len(m.deps)))
	for _, dep := range m.deps {
		out = append(out, fmt.Sprintf("%s- %s", indent, dep))
	}
	keys := dict.Keys(m.tree)
	slices.Sort(keys)
	out = append(out, fmt.Sprintf("Tree: %d", len(m.tree)))
	indent2 := strings.Repeat(indent, 2)
	for _, key := range keys {
		n := m.tree[key]
		numFiles, numFolders := len(n.files), len(n.folders)
		if numFiles == 0 {
			continue // skip if no files
		}
		out = append(out, fmt.Sprintf("%s- %s (%d, %d)", indent, key, numFiles, numFolders))
		for _, file := range n.files {
			out = append(out, fmt.Sprintf("%s* %s", indent2, file))
		}
	}
	return strings.Join(out, "\n")
}

func newModule() *module {
	return &module{
		path: "",
		name: "",
		deps: make([]string, 0),
		tree: make(map[string]*node),
	}
}

func newNode() *node {
	return &node{
		files:   make([]string, 0),
		folders: make([]string, 0),
	}
}
