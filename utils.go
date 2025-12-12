package main

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/roidaradal/fn/number"
	"github.com/roidaradal/fn/str"
)

var (
	startsWith = strings.HasPrefix
	endsWith   = strings.HasSuffix
)

// Split line by space, return part at given index if valid.
// If negative index, gets from the back (-1 is last)
func getLinePart(line string, index int) (string, bool) {
	parts := str.SpaceSplit(line)
	numParts := len(parts)
	if index < 0 {
		index = numParts + index
	}
	if numParts >= index+1 {
		return parts[index], true
	}
	return "", false
}

// Join path by /
func joinPath(path1, path2 string) string {
	return fmt.Sprintf("%s/%s", path1, path2)
}

// Return percentage string
func percentage(num, denom int) string {
	ratio := number.Ratio(num*100, denom)
	return fmt.Sprintf("%.0f%%", ratio)
}

// Return average string
func average(num, denom int) string {
	if denom == 0 {
		return "0"
	}
	ratio := int(number.Ratio(num, denom))
	return number.Comma(ratio)
}

// Sort CountEntries by descending order of counts
// Tie-breaker: alphabetical
func sortDescCount(a, b CountEntry) int {
	score1 := cmp.Compare(b.Value, a.Value)
	if score1 != 0 {
		return score1
	}
	return cmp.Compare(a.Key, b.Key)
}

// Convert package name to node name
func packageToNodeName(name string) string {
	if !startsWith(name, "/") {
		name = "/" + name
	}
	return name
}

// Remove prefix / from node name
func nodeToPackageName(name string) string {
	return str.GuardWith(strings.TrimPrefix(name, "/"), "/")
}
