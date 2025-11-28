package needle

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/ds"
	"github.com/roidaradal/fn/io"
	"github.com/roidaradal/fn/lang"
	"github.com/roidaradal/fn/list"
)

// Build new DepsModule for Go module at path
func NewDepsModule(path string) (*DepsModule, error) {
	// Initialize deps module
	baseMod, err := baseModule(path)
	if err != nil {
		return nil, err
	}
	mod := newDepsModule(baseMod)
	// Apply decorators
	decorators := []func(*DepsModule) error{
		externalDependencies,
		internalDependencies,
		computeDependencyLevels,
	}
	err = applyDecorators(mod, decorators)
	if err != nil {
		return nil, err
	}
	return mod, nil
}

// Process the external package dependencies of each subpackage
func externalDependencies(mod *DepsModule) error {
	type data struct {
		folder  string
		extDeps []string
	}

	// Run concurrently
	cfg := &taskConfig[TreeEntry, data]{
		Task: func(entry TreeEntry) (data, error) {
			var d data
			folder, node := entry.Tuple()
			extDeps, err := packageDependencies(mod.Module, folder, node.Files, false)
			if err != nil {
				return d, err
			}
			return data{folder, extDeps}, nil
		},
		Receive: func(d data) {
			for _, extPkg := range d.extDeps {
				mod.ExternalUsers[extPkg] = append(mod.ExternalUsers[extPkg], d.folder)
			}
		},
	}
	entries := mod.ValidTreeEntries()
	err := runConcurrent(entries, cfg)
	if err != nil {
		return err
	}

	// Sort external users values
	dict.SortValues(mod.ExternalUsers)
	return nil
}

// Process the internal package dependencies of each subpackage
func internalDependencies(mod *DepsModule) error {
	type data struct {
		folder  string
		subDeps []string
	}

	// Run concurrently
	cfg := &taskConfig[TreeEntry, data]{
		Task: func(entry TreeEntry) (data, error) {
			var d data
			folder, node := entry.Tuple()
			subDeps, err := packageDependencies(mod.Module, folder, node.Files, true)
			if err != nil {
				return d, err
			}
			return data{folder, subDeps}, nil
		},
		Receive: func(d data) {
			mod.DependenciesOf[d.folder] = d.subDeps
		},
	}
	entries := mod.ValidTreeEntries()
	err := runConcurrent(entries, cfg)
	if err != nil {
		return err
	}

	// Compute inverse dependency => which package uses it
	mod.InternalUsers = dict.GroupByValueList(mod.DependenciesOf)
	dict.SortValues(mod.InternalUsers)
	return nil
}

// Compute the independent subpackages and dependency levels (dependency DAG)
func computeDependencyLevels(mod *DepsModule) error {
	inbound := mod.InternalUsers   // list of subpackages that use it
	outbound := mod.DependenciesOf // list of subpackages it uses

	// Compute independent subpackages
	q := ds.NewQueue[string](len(outbound))
	for subPkg := range outbound {
		// Independent subpackage = no inbound & outbound
		if len(inbound[subPkg]) == 0 && len(outbound[subPkg]) == 0 {
			mod.IndependentSubs = append(mod.IndependentSubs, subPkg)
		} else {
			q.Enqueue(subPkg)
		}
	}
	slices.Sort(mod.IndependentSubs)

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
	mod.DependencyLevels = dict.GroupByValue(levelOf)
	dict.SortValues(mod.DependencyLevels)

	return nil
}

// Gather dependencies of files in folder path
func packageDependencies(mod *Module, path string, files []string, isInternal bool) ([]string, error) {
	rootFolder := mod.Path + path
	pkgDeps := ds.NewSet[string]()

	// Run concurrently
	cfg := &taskConfig[string, []string]{
		Task: func(filename string) ([]string, error) {
			path := fmt.Sprintf("%s/%s", rootFolder, filename)
			return fileDependencies(mod, path, isInternal)
		},
		Receive: func(deps []string) {
			pkgDeps.AddItems(deps)
		},
	}
	err := runConcurrent(files, cfg)
	if err != nil {
		return nil, err
	}

	// Get the unique list of pkg dependencies and sort it
	deps := pkgDeps.Items()
	slices.Sort(deps)
	return deps, nil
}

// Get list of dependencies from given path file
func fileDependencies(mod *Module, path string, isInternal bool) ([]string, error) {
	if !io.PathExists(path) {
		return nil, fmt.Errorf("file '%s' does not exist", path)
	}
	lines, err := io.ReadNonEmptyLines(path)
	if err != nil {
		return nil, err
	}
	isValid := lang.Ternary(isInternal, isInternalDependency, isExternalDependency)
	deps := make([]string, 0)
	depMode := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "import ") {
			if strings.HasSuffix(line, "(") {
				// toggle depMode on
				depMode = true
			} else {
				// single-line import
				if dep, ok := getLinePart(line, 1); ok {
					if dep, ok := isValid(mod, dep); ok {
						deps = append(deps, dep)
					}
				}
				// end early after processing single-line import
				break
			}
		} else if line == ")" {
			// end early after finishing import block
			break
		} else if depMode {
			// add if valid dependency
			if dep, ok := isValid(mod, line); ok {
				deps = append(deps, dep)
			}
		}
	}
	slices.Sort(deps)
	return deps, nil
}

// Get subpackages that are part of dependency tree
func (mod DepsModule) DependencyPackages() []string {
	return dict.Keys(dict.SwapList(mod.DependencyLevels))
}

// DepsModule string representation
func (mod DepsModule) String() string {
	// Initialize with base Module output
	out := []string{mod.Module.String()}
	// External dependencies
	out = append(out, fmt.Sprintf("ExtDeps: %d", len(mod.ExternalDeps)))
	entries := dict.Entries(mod.ExternalUsers)
	slices.SortFunc(entries, func(a, b dict.Entry[string, []string]) int {
		// Sort by descending order of dependent counts
		return cmp.Compare(len(b.Value), len(a.Value))
	})
	for _, entry := range entries {
		extPkg, users := entry.Tuple()
		out = append(out, fmt.Sprintf("\t%d : %s", len(users), extPkg))
		out = append(out, "\t\t"+strings.Join(users, " "))
	}
	// Compute dependent, independent counts
	totalCount := mod.CountValidNodes()
	indepCount := len(mod.IndependentSubs)
	depCount := totalCount - indepCount
	// Dependency Tree
	out = append(out, fmt.Sprintf("DepSubs: %d / %d", depCount, totalCount))
	maxLength := getMaxLength(mod.DependencyPackages())
	template := fmt.Sprintf("\tL%%d: %%-%ds %%2d | %%d", maxLength)
	template2 := fmt.Sprintf("\t%%%ds | %%s", maxLength+7)
	levels := dict.Keys(mod.DependencyLevels)
	slices.Sort(levels)
	for _, level := range levels {
		for _, subPkg := range mod.DependencyLevels[level] {
			if slices.Contains(mod.IndependentSubs, subPkg) {
				continue
			}
			inCount := len(mod.InternalUsers[subPkg])
			outCount := len(mod.DependenciesOf[subPkg])
			out = append(out, fmt.Sprintf(template, level, subPkg, outCount, inCount))
			for i := range max(inCount, outCount) {
				inboundDep, outboundDep := "", ""
				if i < inCount {
					inboundDep = "-> " + mod.InternalUsers[subPkg][i]
				}
				if i < outCount {
					outboundDep = mod.DependenciesOf[subPkg][i] + " <-"
				}
				out = append(out, fmt.Sprintf(template2, outboundDep, inboundDep))
			}
		}
	}
	// Independent subpackages
	out = append(out, fmt.Sprintf("IndepSubs: %d / %d", indepCount, totalCount))
	if indepCount > 0 {
		out = append(out, "\t"+strings.Join(mod.IndependentSubs, " "))
	}
	return strings.Join(out, "\n")
}
