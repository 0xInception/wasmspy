package main

import (
	"fmt"
	"os"

	"github.com/0xInception/wasmspy/pkg/wasm"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: wasmspy <file.wasm>\n")
		os.Exit(1)
	}

	path := os.Args[1]

	mod, err := wasm.ParseFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing file: %v\n", err)
		os.Exit(1)
	}

	resolved, err := wasm.Resolve(mod)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error resolving module: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(resolved.ToWAT())
}
