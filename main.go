package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/roidaradal/fn/io"
	"github.com/roidaradal/needle/internal/needle"
)

func main() {
	modulePath := getArgs()
	mod, err := needle.BuildModule(modulePath)
	if err != nil {
		log.Fatal(err)
	}
	outputPath, err := needle.BuildReport(mod)
	if err != nil {
		log.Fatal(err)
	}
	outputPath, _ = filepath.Abs(outputPath)

	err = io.OpenFile(outputPath)
	if err != nil {
		log.Fatal(err)
	}
}

// Get module path and output path from command-line args
func getArgs() (modulePath string) {
	args := os.Args[1:]
	numArgs := len(args)
	if numArgs < 1 {
		fmt.Println("Usage: needle <modulePath>")
		os.Exit(1)
	}
	modulePath = args[0]
	return modulePath
}
