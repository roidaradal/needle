package main

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/roidaradal/fn/conk"
	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/ds"
	"github.com/roidaradal/fn/io"
	"github.com/roidaradal/fn/list"
)

// Process the internal / external dependencies of each subpackage in the module
func modPackageDependencies(mod *Module) error {
	type data struct {
		name string
		deps map[string]bool // isInternal
	}

	// Run concurrently
	task := func(entry NodeEntry) (data, error) {
		var d data
		name, node := entry.Tuple()
		deps, err := packageDependencies(mod, name, node.Files)
		if err != nil {
			return d, err
		}
		return data{name, deps}, nil
	}
	onReceive := func(d data) {
		for dep, isInternal := range d.deps {
			if isInternal {
				mod.Deps.Of[d.name] = append(mod.Deps.Of[d.name], dep)
			} else {
				mod.ExternalUsers[dep] = append(mod.ExternalUsers[dep], d.name)
			}
		}
	}
	entries := mod.packageNodeEntries()
	err := conk.Tasks(entries, task, onReceive)
	if err != nil {
		return err
	}

	// Compute inverse dependency => which package uses it
	mod.Deps.InternalUsers = dict.GroupByValueList(mod.Deps.Of)
	dict.SortValues(mod.Deps.InternalUsers)
	dict.SortValues(mod.Deps.ExternalUsers)
	dict.SortValues(mod.Deps.Of)
	return nil
}

// Gather dependency data of files in folder path
func packageDependencies(mod *Module, path string, files []string) (map[string]bool, error) {
	folder := mod.Path + path
	pkgDeps := make(map[string]bool)

	// Run concurrently
	task := func(filename string) (map[string]bool, error) {
		path := filepath.Join(folder, filename)
		return fileDependencies(mod, path)
	}
	onReceive := func(deps map[string]bool) {
		maps.Copy(pkgDeps, deps)
	}
	err := conk.Tasks(files, task, onReceive)
	if err != nil {
		return nil, err
	}
	return pkgDeps, nil
}

// Get list of dependencies from give file path
func fileDependencies(mod *Module, path string) (map[string]bool, error) {
	if !io.PathExists(path) {
		return nil, fmt.Errorf("file %q does not exist", path)
	}
	lines, err := io.ReadNonEmptyLines(path)
	if err != nil {
		return nil, err
	}
	candidates := make([]string, 0)
	inDepMode := false
	for _, line := range lines {
		if startsWith(line, "import ") {
			if endsWith(line, "(") {
				inDepMode = true // toggle DepMode on
			} else {
				// single line import
				if dep, ok := getLinePart(line, 1); ok {
					candidates = append(candidates, dep)
				}
			}
		} else if inDepMode {
			if line == ")" {
				inDepMode = false // toggle DepMode off
				continue
			}
			candidates = append(candidates, line)
		}
	}
	deps := make(map[string]bool)
	for _, candidate := range candidates {
		if internalDep, ok := isInternalDependency(mod, candidate); ok {
			deps[internalDep] = true
			continue
		}
		if externalDep, ok := isExternalDependency(mod, candidate); ok {
			deps[externalDep] = false
		}
	}
	return deps, nil
}

// Check if package is internal dependency,
// return the processed name
func isInternalDependency(mod *Module, dep string) (string, bool) {
	dep = strings.Trim(dep, "\"")
	isInternal := startsWith(dep, mod.Name)
	if isInternal {
		dep = strings.TrimPrefix(dep, mod.Name)
		// TODO: replace with str.GuardWith
		if dep == "" {
			dep = "/"
		}
	}
	return dep, isInternal
}

// Check if package is external dependency,
// return the processed name
func isExternalDependency(mod *Module, dep string) (string, bool) {
	dep = strings.Trim(dep, "\"")
	for extPkg := range mod.Deps.ExternalUsers {
		if startsWith(dep, extPkg) {
			return extPkg, true
		}
	}
	return dep, false
}

// Process the independent subpackages and dependency levels (dependency DAG)
func computeDependencyLevels(mod *Module) error {
	inbound := mod.Deps.InternalUsers // list of subpackages that use it
	outbound := mod.Deps.Of           // list of subpackages it uses

	// Compute independent subpackages
	// Add depedency packages to queue
	q := ds.NewQueue[string]()
	for subPkg, node := range mod.Nodes {
		if node.FileCount() == 0 {
			continue // skip no files
		}
		// Independent subpackage = no inbound & outbound
		if len(inbound[subPkg]) == 0 && len(outbound[subPkg]) == 0 {
			mod.Deps.Independent = append(mod.Deps.Independent, subPkg)
		} else {
			q.Enqueue(subPkg)
		}
	}
	slices.Sort(mod.Deps.Independent)

	// Compute tree subpackage levels
	levelOf := make(dict.IntMap)
	for q.NotEmpty() {
		subPkg, _ := q.Dequeue()
		if len(outbound[subPkg]) == 0 {
			// no dependency = level 0
			levelOf[subPkg] = 0
			continue
		}
		isComplete := list.All(outbound[subPkg], func(dep string) bool {
			return dict.HasKey(levelOf, dep)
		})
		if isComplete {
			// All dependencies have levels = get max + 1
			depLevels := list.Translate(outbound[subPkg], levelOf)
			levelOf[subPkg] = slices.Max(depLevels) + 1
			continue
		}
		// Incomplete, put back in queue
		q.Enqueue(subPkg)
	}
	mod.Deps.Levels = dict.GroupByValue(levelOf)
	dict.SortValues(mod.Deps.Levels)

	return nil
}
