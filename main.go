package main

import (
	"fmt"
	"log"
	"os"
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
	err = internalDependency(mod)
	if err != nil {
		log.Fatal(err)
	}

	// External dependency
	err = externalDependency(mod)
	if err != nil {
		log.Fatal(err)
	}

	// Split internal dependencies => independent, levels
	splitIndependentTree(mod)

	// Display text report
	textReport(mod)
}
