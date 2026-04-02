package main

import (
	"fmt"
	"os"

	"github.com/eminbekov/fiber-v3-template/internal/generate"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	var runError error
	switch os.Args[1] {
	case "migration":
		runError = generate.Migration(os.Args[2:])
	case "resource":
		runError = generate.Resource(os.Args[2:])
	default:
		printUsage()
		runError = fmt.Errorf("unknown subcommand: %s", os.Args[1])
	}

	if runError != nil {
		fmt.Fprintf(os.Stderr, "generate: %v\n", runError)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: go run ./cmd/generate <subcommand> [options]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Subcommands:")
	fmt.Fprintln(os.Stderr, "  migration <name>   Create next sequential migration pair")
	fmt.Fprintln(os.Stderr, "  resource <name>    Scaffold domain CRUD resource (implemented in phase 17)")
}
