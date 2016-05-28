package main

import (
	"fmt"
	"os"

	"github.com/coreos/mantle/update/generator"
)

func main() {
	var g generator.Generator

	part, err := generator.FullUpdate(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := g.Partition(part); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := g.Write("out.gz"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
