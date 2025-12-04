package main

import (
	"fmt"
	"log"
	"os"

	"github.com/roidaradal/fn/io"
)

func main() {
	modulePath, outputPath := getArgs()
	mod, err := BuildModule(modulePath)
	if err != nil {
		log.Fatal(err)
	}
	err = BuildReport(mod, outputPath)
	if err != nil {
		log.Fatal(err)
	}
	err = io.OpenFile(outputPath)
	if err != nil {
		log.Fatal(err)
	}
}

// Get module path and output path from command-line args
func getArgs() (modulePath, outputPath string) {
	args := os.Args[1:]
	numArgs := len(args)
	if numArgs < 1 {
		fmt.Println("Usage: needle <modulePath> (<outputPath>)")
		os.Exit(1)
	}
	modulePath = args[0]
	outputPath = "needle.html"
	if numArgs >= 2 {
		outputPath = args[1]
	}
	if !endsWith(outputPath, ".html") {
		fmt.Println("Output path needs to be a .html file")
		os.Exit(1)
	}
	return modulePath, outputPath
}
