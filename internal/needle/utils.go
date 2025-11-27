package needle

import (
	"strings"

	"github.com/roidaradal/fn/list"
	"github.com/roidaradal/fn/str"
)

// Check if line ends with // indirect
func isDirectDependency(line string) bool {
	return !strings.HasSuffix(line, "// indirect")
}

// Check if directory path doesn't start with dot, underscore, dash
func isPublicFolder(name string) bool {
	prefixes := []string{".", "_", "-"}
	return list.AllTrue(list.Map(prefixes, func(prefix string) bool {
		return !strings.HasPrefix(name, prefix)
	}))
}

// Split the line by space, and return the part at given index
func getLinePart(line string, index int) (string, bool) {
	parts := str.SpaceSplit(line)
	if len(parts) >= index+1 {
		return parts[index], true
	}
	return "", false
}

// Check if package is an internal dependency,
// Return the processed dependency name
func isInternalDependency(mod *Module, dep string) (string, bool) {
	dep = strings.Trim(dep, "\"")
	isInternal := strings.HasPrefix(dep, mod.Name)
	if isInternal {
		dep = strings.TrimPrefix(dep, mod.Name)
		if dep == "" {
			dep = "/"
		}
	}
	return dep, isInternal
}

// Check if package is an external dependency,
// Return the processed dependency name
func isExternalDependency(mod *Module, dep string) (string, bool) {
	dep = strings.Trim(dep, "\"")
	for _, extPkg := range mod.ExternalDeps {
		if strings.HasPrefix(dep, extPkg) {
			return extPkg, true
		}
	}
	return dep, false
}
