package needle

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/roidaradal/fn/lang"
	"github.com/roidaradal/fn/list"
	"github.com/roidaradal/fn/number"
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

// Apply list of decorator functions to mod object
func applyDecorators[T any](mod *T, decorators []func(*T) error) error {
	for _, decorator := range decorators {
		err := decorator(mod)
		if err != nil {
			return err
		}
	}
	return nil
}

// Create package name from path
func getPackageName(path string) string {
	return str.GuardDot(strings.TrimPrefix(path, "/"))
}

// Extract filename from full path
func getFilename(path string) string {
	parts := str.CleanSplit(path, "/")
	return parts[len(parts)-1]
}

// Get FileType of file path
func getFileType(path string) FileType {
	return lang.Ternary(strings.HasSuffix(path, "_test.go"), FILE_TEST, FILE_CODE)
}

// Get max length from list of strings
func getMaxLength(items []string) int {
	return slices.Max(list.Map(items, str.Length))
}

// Return percentage string
func percentage(num, denom int) string {
	ratio := number.Ratio(num*100, denom)
	return fmt.Sprintf("%.0f%%", ratio)
}

// Sort CountEntries by descending order of counts
func sortDescCount(a, b CountEntry) int {
	return cmp.Compare(b.Value, a.Value)
}

// Create breakdown header
func breakdownHeader() string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
		str.Center("Code", 8),
		str.Center("Error", 8),
		str.Center("Head", 8),
		str.Center("Comment", 8),
		str.Center("Space", 8),
		str.Center("Total", 10),
		str.Center("Packages", 20),
	)
}

// Create breakdown row
func breakdownRow(code, err, head, comment, space, total int, pkg string) string {
	return fmt.Sprintf("%8s|%8s|%8s|%8s|%8s|%10s| %s",
		number.Comma(code),
		number.Comma(err),
		number.Comma(head),
		number.Comma(comment),
		number.Comma(space),
		number.Comma(total),
		pkg,
	)
}

// Create breakdown percentage with 1 value
func breakdownRowString(code, err, head, comment, space, total, pkg string) string {
	return fmt.Sprintf("%8s|%8s|%8s|%8s|%8s|%10s| %s",
		code,
		err,
		head,
		comment,
		space,
		total,
		pkg,
	)
}

// Create percentage details
func percentDetails(num, denom1, denom2 int) string {
	if num == 0 {
		return ""
	}
	p1, p2 := "-", "-"
	if denom1 > 0 {
		p1 = percentage(num, denom1)
	}
	if denom2 > 0 {
		p2 = percentage(num, denom2)
	}
	return fmt.Sprintf("%s,%s", p1, p2)
}
