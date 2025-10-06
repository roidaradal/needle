package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Usage: needle <modulePath>")
		return
	}
	modulePath := args[0]
	fmt.Println(modulePath)
}
