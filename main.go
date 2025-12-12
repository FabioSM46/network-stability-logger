// Package main is the entry point for the network stability monitor.
package main

import (
	"fmt"
	"os"

	"github.com/FabioSM46/network-stability-logger/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
