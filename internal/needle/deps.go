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
	"github.com/roidaradal/fn/str"
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
	for _, decorator := range decorators {
		err = decorator(mod)
		if err != nil {
			return nil, err
		}
	}
	return mod, nil
}

// Process the external package dependencies of each subpackage
func externalDependencies(mod *DepsModule) error {
	for folder, node := range mod.Tree {
		if len(node.Files) == 0 {
			continue // skip if no files
		}
		extDeps, err := packageDependencies(mod.Module, folder, node.Files, false)
		if err != nil {
			return err
		}
		for _, extPkg := range extDeps {
			mod.ExternalUsers[extPkg] = append(mod.ExternalUsers[extPkg], folder)
		}
	}
	dict.SortValues(mod.ExternalUsers)
	return nil
}

// Process the internal package dependencies of each subpackage
func internalDependencies(mod *DepsModule) error {
	for folder, node := range mod.Tree {
		if len(node.Files) == 0 {
			continue // skip if no files
		}
		subDeps, err := packageDependencies(mod.Module, folder, node.Files, true)
		if err != nil {
			return err
		}
		mod.DependenciesOf[folder] = subDeps
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
	pkgDeps := ds.NewSet[string]()
	rootFolder := mod.Path + path
	for _, filename := range files {
		path := fmt.Sprintf("%s/%s", rootFolder, filename)
		deps, err := fileDependencies(mod, path, isInternal)
		if err != nil {
			return nil, err
		}
		pkgDeps.AddItems(deps)
	}
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
func (m DepsModule) DependencyPackages() []string {
	return dict.Keys(dict.SwapList(m.DependencyLevels))
}

// DepsModule string representation
func (m DepsModule) String() string {
	// Initialize with base Module output
	out := []string{m.Module.String()}
	// External dependencies
	out = append(out, fmt.Sprintf("ExtDeps: %d", len(m.ExternalDeps)))
	entries := dict.Entries(m.ExternalUsers)
	slices.SortFunc(entries, func(a, b dict.Entry[string, []string]) int {
		// Sort by descending order of dependent counts
		return cmp.Compare(len(b.Value), len(a.Value))
	})
	for _, entry := range entries {
		extPkg, users := entry.Key, entry.Value
		out = append(out, fmt.Sprintf("\t%d : %s", len(users), extPkg))
		out = append(out, "\t\t"+strings.Join(users, " "))
	}
	// Compute dependent, independent counts
	totalCount := m.CountValidNodes()
	indepCount := len(m.IndependentSubs)
	depCount := totalCount - indepCount
	// Dependency Tree
	out = append(out, fmt.Sprintf("DepSubs: %d / %d", depCount, totalCount))
	maxLength := slices.Max(list.Map(m.DependencyPackages(), str.Length))
	template := fmt.Sprintf("\tL%%d: %%-%ds %%2d | %%d", maxLength)
	template2 := fmt.Sprintf("\t%%%ds | %%s", maxLength+7)
	levels := dict.Keys(m.DependencyLevels)
	slices.Sort(levels)
	for _, level := range levels {
		for _, subPkg := range m.DependencyLevels[level] {
			if slices.Contains(m.IndependentSubs, subPkg) {
				continue
			}
			inCount := len(m.InternalUsers[subPkg])
			outCount := len(m.DependenciesOf[subPkg])
			out = append(out, fmt.Sprintf(template, level, subPkg, outCount, inCount))
			for i := range max(inCount, outCount) {
				inboundDep, outboundDep := "", ""
				if i < inCount {
					inboundDep = m.InternalUsers[subPkg][i]
				}
				if i < outCount {
					outboundDep = m.DependenciesOf[subPkg][i]
				}
				out = append(out, fmt.Sprintf(template2, outboundDep, inboundDep))
			}
		}
	}
	// Independent subpackages
	out = append(out, fmt.Sprintf("IndepSubs: %d / %d", indepCount, totalCount))
	if indepCount > 0 {
		out = append(out, "\t"+strings.Join(m.IndependentSubs, " "))
	}
	return strings.Join(out, "\n")
}
