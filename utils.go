package main

import (
	"fmt"
	"strings"

	"github.com/roidaradal/fn/str"
)

var (
	startsWith = strings.HasPrefix
	endsWith   = strings.HasSuffix
)

// Split line by space, return part at given index if valid
func getLinePart(line string, index int) (string, bool) {
	parts := str.SpaceSplit(line)
	if len(parts) >= index+1 {
		return parts[index], true
	}
	return "", false
}

// Join path by /
func joinPath(path1, path2 string) string {
	return fmt.Sprintf("%s/%s", path1, path2)
}
