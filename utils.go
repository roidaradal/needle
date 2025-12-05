package main

import (
	"fmt"
	"strings"

	"github.com/roidaradal/fn/list"
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

// Wrap string by li tags
func wrapLiTags(text string) string {
	return fmt.Sprintf("<li>%s</li>", text)
}

// Create list items string
func listItems(items []string) string {
	return strings.Join(list.Map(items, wrapLiTags), "")
}

// TODO: update with dict.UpdateCounts
func dictUpdateCounts[K comparable](oldCounter, newCounter map[K]int) {
	for key, count := range newCounter {
		oldCounter[key] += count
	}
}
