package main

import (
	"fmt"
	"os"

	"github.com/0xInception/wasmspy/pkg/decompile"
	"github.com/0xInception/wasmspy/pkg/wasm"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "wat":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "usage: wasmspy wat <file.wasm>\n")
			os.Exit(1)
		}
		cmdWAT(os.Args[2])

	case "decompile":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "usage: wasmspy decompile <file.wasm> [func_name]\n")
			os.Exit(1)
		}
		funcName := ""
		if len(os.Args) >= 4 {
			funcName = os.Args[3]
		}
		cmdDecompile(os.Args[2], funcName)

	case "callgraph":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "usage: wasmspy callgraph <file.wasm>\n")
			os.Exit(1)
		}
		cmdCallGraph(os.Args[2])

	case "info":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "usage: wasmspy info <file.wasm>\n")
			os.Exit(1)
		}
		cmdInfo(os.Args[2])

	case "help", "-h", "--help":
		usage()

	default:
		if _, err := os.Stat(cmd); err == nil {
			cmdWAT(cmd)
		} else {
			fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
			usage()
			os.Exit(1)
		}
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `wasmspy - WebAssembly inspector and decompiler

usage: wasmspy <command> <file.wasm> [options]

commands:
  wat        output WAT format (default)
  decompile  decompile to pseudocode
  callgraph  show function call graph
  info       show module information
  help       show this help

examples:
  wasmspy wat module.wasm
  wasmspy decompile module.wasm
  wasmspy decompile module.wasm main
  wasmspy callgraph module.wasm
`)
}

func loadModule(path string) *wasm.ResolvedModule {
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

	return resolved
}

func cmdWAT(path string) {
	module := loadModule(path)
	fmt.Println(module.ToWAT())
}

func cmdDecompile(path, funcName string) {
	module := loadModule(path)

	if funcName != "" {
		fn := module.GetFunctionByName(funcName)
		if fn == nil {
			fmt.Fprintf(os.Stderr, "function not found: %s\n", funcName)
			os.Exit(1)
		}
		fmt.Println(decompile.Decompile(fn, module))
	} else {
		fmt.Println(decompile.DecompileModule(module))
	}
}

func cmdCallGraph(path string) {
	module := loadModule(path)
	cg := decompile.BuildCallGraph(module)

	output := cg.String(module)
	if output == "" {
		fmt.Println("(no calls between functions)")
	} else {
		fmt.Print(output)
	}

	roots := cg.Roots(module)
	if len(roots) > 0 {
		fmt.Printf("\nentry points: ")
		for i, idx := range roots {
			if i > 0 {
				fmt.Print(", ")
			}
			if fn := module.GetFunction(idx); fn != nil && fn.Name != "" {
				fmt.Print(fn.Name)
			} else {
				fmt.Printf("func_%d", idx)
			}
		}
		fmt.Println()
	}
}

func cmdInfo(path string) {
	module := loadModule(path)

	fmt.Printf("version: %d\n", module.Version)
	fmt.Printf("functions: %d\n", len(module.Functions))
	fmt.Printf("types: %d\n", len(module.Types))
	fmt.Printf("tables: %d\n", len(module.Tables))
	fmt.Printf("memories: %d\n", len(module.Memories))
	fmt.Printf("globals: %d\n", len(module.Globals))

	imports := 0
	for _, fn := range module.Functions {
		if fn.Imported {
			imports++
		}
	}
	fmt.Printf("imports: %d\n", imports)
	fmt.Printf("exports: %d\n", len(module.Exports))

	if len(module.Exports) > 0 {
		fmt.Println("\nexports:")
		for _, exp := range module.Exports {
			fmt.Printf("  %s (%s)\n", exp.Name, exportKind(exp.Kind))
		}
	}

	if imports > 0 {
		fmt.Println("\nimports:")
		for _, imp := range module.Imports {
			fmt.Printf("  %s.%s (%s)\n", imp.Module, imp.Name, importKind(imp.Kind))
		}
	}
}

func exportKind(k wasm.ExportKind) string {
	switch k {
	case wasm.ExportFunc:
		return "func"
	case wasm.ExportTable:
		return "table"
	case wasm.ExportMemory:
		return "memory"
	case wasm.ExportGlobal:
		return "global"
	}
	return "unknown"
}

func importKind(k wasm.ImportKind) string {
	switch k {
	case wasm.ImportFunc:
		return "func"
	case wasm.ImportTable:
		return "table"
	case wasm.ImportMemory:
		return "memory"
	case wasm.ImportGlobal:
		return "global"
	}
	return "unknown"
}
