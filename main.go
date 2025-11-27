package main

import (
	"fmt"
	"log"
	"os"

	"github.com/roidaradal/needle/internal/needle"
)

const usage string = "Usage: needle <deps|viz|stats|api> <path>"

func main() {
	option, path := getArgs()
	switch option {
	case "deps":
		mod, err := needle.NewDepsModule(path)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(mod)
	default:
		fmt.Println(usage)
	}
}

func getArgs() (option, path string) {
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}
	option, path = args[0], args[1]
	return option, path
}
