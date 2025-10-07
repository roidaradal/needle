package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Usage: needle <modulePath>")
		return
	}

	folder := args[0]
	err := getModuleInfo(folder)
	if err != nil {
		log.Fatal(err)
	}
}
