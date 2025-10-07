package main

import (
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/roidaradal/fn/dict"
)

func main() {
	// Get folder from args
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Usage: needle <modulePath>")
		return
	}
	folder := args[0]

	// Get module info
	mod, err := getModuleInfo(folder)
	if err != nil {
		log.Fatal(err)
	}

	// Internal dependency
	inDep, err := internalDependency(mod)
	if err != nil {
		log.Fatal(err)
	}
	independent, levels := splitIndependentTree(inDep)
	fmt.Println("Independent:", len(independent), independent)
	fmt.Println("Levels:", len(levels))
	for level := range len(levels) {
		fmt.Println(level, levels[level])
	}

	// External dependency
	extDep, err := externalDependency(mod)
	if err != nil {
		log.Fatal(err)
	}
	keys := dict.Keys(extDep)
	slices.Sort(keys)
	fmt.Println("External:", len(keys))
	for _, k := range keys {
		fmt.Println(k, extDep[k])
	}
}
