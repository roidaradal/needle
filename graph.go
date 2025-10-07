package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/roidaradal/fn"
	"github.com/roidaradal/fn/check"
	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/ds"
	"github.com/roidaradal/fn/io"
	"github.com/roidaradal/fn/str"
)

// Split the independent packages vs tree packages, from the internal dependency
func splitIndependentTree(out dict.StringListMap) ([]string, map[int][]string) {
	in := make(dict.StringListMap)
	for pkg, deps := range out {
		for _, dep := range deps {
			in[dep] = append(in[dep], pkg)
		}
	}

	independent := make([]string, 0)
	q := ds.NewQueue[string]()
	for pkg := range out {
		// Independent packages = no in & out
		if len(in[pkg]) == 0 && len(out[pkg]) == 0 {
			independent = append(independent, pkg)
		} else {
			q.Enqueue(pkg)
		}
	}
	slices.Sort(independent)

	levelOf := make(map[string]int)
	for !q.IsEmpty() {
		pkg, _ := q.Dequeue()
		if len(out[pkg]) == 0 {
			// No dependency = level 0
			levelOf[pkg] = 0
			continue
		}
		complete := check.All(out[pkg], func(dep string) bool {
			return dict.HasKey(levelOf, dep)
		})
		if complete {
			// All dependencies have computed levels = get max + 1
			depLevels := fn.Translate(out[pkg], levelOf)
			levelOf[pkg] = slices.Max(depLevels) + 1
			continue
		}
		// Incomplete, put back in queue
		q.Enqueue(pkg)
	}

	// Group the packages by level, sort the packages per level
	levels := make(map[int][]string)
	for pkg, level := range levelOf {
		levels[level] = append(levels[level], pkg)
	}
	for level, pkgs := range levels {
		slices.Sort(pkgs)
		levels[level] = pkgs
	}

	return independent, levels
}

// Compute internal dependency of each package in mod.tree
func internalDependency(mod *module) (dict.StringListMap, error) {
	dep := make(dict.StringListMap)
	for folder, node := range mod.tree {
		if len(node.files) == 0 {
			continue // skip if no files
		}
		deps, err := packageDependency(mod, folder, node.files, true)
		if err != nil {
			return nil, err
		}
		dep[folder] = deps
	}
	return dep, nil
}

// Compute dependency of packages to external packages
func externalDependency(mod *module) (dict.StringListMap, error) {
	dep := make(dict.StringListMap)
	for folder, node := range mod.tree {
		if len(node.files) == 0 {
			continue // skip if no files
		}
		deps, err := packageDependency(mod, folder, node.files, false)
		if err != nil {
			return nil, err
		}
		for _, extPkg := range deps {
			dep[extPkg] = append(dep[extPkg], folder)
		}
	}
	return dep, nil
}

// Gather dependency of files in folder
func packageDependency(mod *module, folder string, files []string, internal bool) ([]string, error) {
	pkgDeps := ds.NewSet[string]()
	for _, filename := range files {
		path := mod.path + folder + filename
		deps, err := fileDependency(mod, path, internal)
		if err != nil {
			return nil, err
		}
		for _, dep := range deps {
			pkgDeps.Add(dep)
		}
	}
	return pkgDeps.Items(), nil
}

// Get list of dependencies of one file
func fileDependency(mod *module, path string, internal bool) ([]string, error) {
	depCheck := fn.Ternary(internal, isInternal, isExternal)
	if !io.PathExists(path) {
		return nil, fmt.Errorf("file '%s' does not exist", path)
	}

	lines, err := io.ReadLines(path)
	if err != nil {
		return nil, err
	}

	deps := make([]string, 0)
	depMode := false
	for _, line := range lines {
		if strings.HasPrefix(line, "import ") {
			if strings.HasSuffix(line, "(") {
				depMode = true // toggle dep mode on
			} else {
				// single line import
				parts := str.SpaceSplit(line)
				if len(parts) >= 2 {
					dep, ok := depCheck(mod, parts[1])
					if ok {
						deps = append(deps, dep)
					}
				}
				break // end early after processing single-line import
			}
		} else if line == ")" {
			break // end early after finishing import block
		} else if depMode {
			dep, ok := depCheck(mod, line)
			if ok {
				deps = append(deps, dep)
			}
		}
	}
	return deps, nil
}

// Check if package dependency is an internal dependency
func isInternal(mod *module, dep string) (string, bool) {
	dep = strings.Trim(dep, "\"")
	internal := strings.HasPrefix(dep, mod.name)
	if internal {
		if dep == mod.name {
			dep = "/" // root
		} else {
			dep = strings.TrimPrefix(dep, mod.name) + "/"
		}
	}
	return dep, internal
}

// Check if package dependency is an external dependency
func isExternal(mod *module, dep string) (string, bool) {
	dep = strings.Trim(dep, "\"")
	external := false
	for _, extDep := range mod.deps {
		if strings.HasPrefix(dep, extDep) {
			dep = extDep
			external = true
			break
		}
	}
	return dep, external
}
