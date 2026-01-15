package needle

import (
	"fmt"
	"slices"
	"strings"

	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/ds"
	"github.com/roidaradal/fn/list"
	"github.com/roidaradal/fn/str"
)

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

// Compute the dependency graph layout
func computeDependencyLayout(mod *Module) error {
	mod.Deps.Nodes = make(dict.StringMap)
	mod.Deps.Edges = make([]string, 0)
	const columnWidth, rowHeight = 100, 100
	const cellRadius = 50

	// Add edge list
	for pkg, deps := range mod.Deps.Of {
		pkg = nodeToPackageName(pkg)
		for _, dependency := range deps {
			dependency = nodeToPackageName(dependency)
			mod.Deps.Edges = append(mod.Deps.Edges, fmt.Sprintf("'%s-%s'", pkg, dependency))
		}
	}

	numLevels := len(mod.Deps.Levels)
	for col := range numLevels {
		level := numLevels - col - 1
		isSink := level == 0
		x := (columnWidth * col) + cellRadius
		offset := 50 * (col % 2)
		for row, pkg := range mod.Deps.Levels[level] {
			pkg = nodeToPackageName(pkg)
			y := (rowHeight * row) + cellRadius + offset
			mod.Deps.Nodes[pkg] = fmt.Sprintf("{x : %d, y: %d, sink: %v}", x, y, isSink)
		}
	}

	return nil
}

// Add file dependency
func (f *File) addDependency(mod *Module, dep string) {
	if internalDep, ok := isInternalDependency(mod, dep); ok {
		f.Deps[internalDep] = true
		return
	}
	if externalDep, ok := isExternalDependency(mod, dep); ok {
		f.Deps[externalDep] = false
		return
	}
}

// Check if package is internal dependency,
// return the processed name
func isInternalDependency(mod *Module, dep string) (string, bool) {
	dep = strings.Trim(dep, "\"")
	isInternal := startsWith(dep, mod.Name)
	if isInternal {
		dep = str.GuardWith(strings.TrimPrefix(dep, mod.Name), "/")
	}
	return dep, isInternal
}

// Check if package is external dependency,
// return the processed name
func isExternalDependency(mod *Module, dep string) (string, bool) {
	dep = strings.Trim(dep, "\"")
	for _, extPkg := range mod.Deps.External {
		if startsWith(dep, extPkg) {
			return extPkg, true
		}
	}
	return dep, false
}
