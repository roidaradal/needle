package main

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/roidaradal/fn/dict"
)

const indent string = "  "

var separator = strings.Repeat("-----", 10)

func textReport(mod *module) {
	indent3 := strings.Repeat(indent, 3)
	indent4 := strings.Repeat(indent, 4)

	// Module info
	fmt.Println(separator)
	fmt.Println("Module:", mod.name)

	// External dependencies
	fmt.Println(separator)
	fmt.Println("External:", len(mod.externals))
	entries := dict.Entries(mod.externals)
	slices.SortFunc(entries, func(a, b dict.Entry[string, []string]) int {
		// Sort by descending order of dependents
		return cmp.Compare(len(b.Value), len(a.Value))
	})
	for _, entry := range entries {
		extPkg, dependents := entry.Key, entry.Value
		fmt.Printf("%s%-3d%s\n", indent, len(dependents), extPkg)
		for _, dep := range dependents {
			fmt.Printf("%s%s\n", indent3, dep)
		}
	}

	// Internal dependencies
	fmt.Println(separator)
	fmt.Println("Levels:", len(mod.levels))
	for level := range len(mod.levels) {
		for _, pkg := range mod.levels[level] {
			fmt.Printf("%s%-3d%s\n", indent, level, pkg)
			numUsers := len(mod.users[pkg])
			if numUsers > 0 {
				fmt.Printf("%sUser: %d\n", indent3, numUsers)
				for _, user := range mod.users[pkg] {
					fmt.Printf("%s%s\n", indent4, user)
				}
			}
			numDeps := len(mod.dependencies[pkg])
			if numDeps > 0 {
				fmt.Printf("%sDeps: %d\n", indent3, numDeps)
				for _, dep := range mod.dependencies[pkg] {
					fmt.Printf("%s%s\n", indent4, dep)
				}
			}
		}
	}

	// Independent packages
	fmt.Println(separator)
	fmt.Println("Independent:", len(mod.independent))
	for _, pkg := range mod.independent {
		fmt.Printf("%s%s\n", indent, pkg)
	}

	fmt.Println(separator)
}
