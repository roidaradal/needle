package main

import (
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/roidaradal/needle/internal/needle"
)

const usage string = "Usage: needle <deps|stats> <path> (--compact)"

func main() {
	cfg := getArgs()
	switch cfg.Option {
	case "deps":
		mod, err := needle.NewDepsModule(cfg)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(mod)
	case "stats":
		mod, err := needle.NewStatsModule(cfg)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(mod)
	default:
		fmt.Println(usage)
	}
}

func getArgs() *needle.Config {
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}
	option, path := args[0], args[1]
	return &needle.Config{
		Option:    option,
		Path:      path,
		IsCompact: slices.Contains(args, "--compact"),
	}
}
