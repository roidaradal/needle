package needle

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/roidaradal/fn/lang"
	"github.com/roidaradal/fn/list"
	"github.com/roidaradal/fn/number"
	"github.com/roidaradal/fn/str"
	"golang.org/x/sync/errgroup"
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

// Config for concurrent tasks
type taskConfig[T, D any] struct {
	Task    func(T) (D, error)
	Receive func(D)
}

// Run concurrent task for each item
func runConcurrent[T, D any](items []T, cfg *taskConfig[T, D]) error {
	// Run goroutine for each item
	var eg errgroup.Group
	dataCh := make(chan D, len(items))
	for _, item := range items {
		eg.Go(func() error {
			result, err := cfg.Task(item)
			if err != nil {
				return err
			}
			dataCh <- result
			return nil
		})
	}

	// Wait for errgroup and close data channel
	var finalErr error
	go func() {
		finalErr = eg.Wait()
		close(dataCh)
	}()

	// Receive data from data channel
	for result := range dataCh {
		cfg.Receive(result)
	}
	return finalErr
}

// Sort CountEntries by descending order of counts
func sortDescCount(a, b CountEntry) int {
	return cmp.Compare(b.Value, a.Value)
}

// Get indentation string
func getIndentation(rawText string) string {
	suffix := strings.TrimLeftFunc(rawText, unicode.IsSpace)
	return strings.TrimSuffix(rawText, suffix)
}

// Get trailing whitespace
func getTrailingWhitespace(rawText string) string {
	prefix := strings.TrimRightFunc(rawText, unicode.IsSpace)
	return strings.TrimPrefix(rawText, prefix)
}

// TODO: replace with str.Center string
func strCenter(text string, width int) string {
	padCount := width - len(text)
	if padCount <= 0 {
		return text
	}
	pad1 := padCount / 2
	pad2 := padCount - pad1
	return fmt.Sprintf("%s%s%s", strings.Repeat(" ", pad2), text, strings.Repeat(" ", pad1))
}
