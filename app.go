package main

import (
	"context"
	"fmt"

	"github.com/0xInception/wasmspy/pkg/decompile"
	"github.com/0xInception/wasmspy/pkg/wasm"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx     context.Context
	modules map[string]*wasm.ResolvedModule
}

func NewApp() *App {
	return &App{
		modules: make(map[string]*wasm.ResolvedModule),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) onDomReady(ctx context.Context) {
	runtime.OnFileDrop(ctx, func(x, y int, paths []string) {
		for _, path := range paths {
			if len(path) > 5 && path[len(path)-5:] == ".wasm" {
				runtime.EventsEmit(ctx, "filedrop", path)
				return
			}
		}
	})
}

func (a *App) OpenFileDialog() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open WebAssembly Module",
		Filters: []runtime.FileFilter{
			{DisplayName: "WebAssembly Files", Pattern: "*.wasm"},
		},
	})
	return path, err
}

type ModuleInfo struct {
	Functions []FunctionInfo `json:"functions"`
	Exports   []ExportInfo   `json:"exports"`
	Memories  []MemoryInfo   `json:"memories"`
	Tables    []TableInfo    `json:"tables"`
	Globals   []GlobalInfo   `json:"globals"`
}

type MemoryInfo struct {
	Index  int  `json:"index"`
	Min    uint32 `json:"min"`
	Max    uint32 `json:"max"`
	HasMax bool   `json:"hasMax"`
}

type TableInfo struct {
	Index  int  `json:"index"`
	Min    uint32 `json:"min"`
	Max    uint32 `json:"max"`
	HasMax bool   `json:"hasMax"`
}

type GlobalInfo struct {
	Index   int    `json:"index"`
	Type    string `json:"type"`
	Mutable bool   `json:"mutable"`
}

type FunctionInfo struct {
	Index    uint32 `json:"index"`
	Name     string `json:"name"`
	Imported bool   `json:"imported"`
	Params   int    `json:"params"`
	Results  int    `json:"results"`
}

type ExportInfo struct {
	Name  string `json:"name"`
	Kind  string `json:"kind"`
	Index uint32 `json:"index"`
}

func (a *App) LoadModuleFromPath(path string) (*ModuleInfo, error) {
	mod, err := wasm.ParseFile(path)
	if err != nil {
		return nil, err
	}
	return a.loadModule(path, mod)
}

func (a *App) loadModule(path string, mod *wasm.Module) (*ModuleInfo, error) {
	resolved, err := wasm.Resolve(mod)
	if err != nil {
		return nil, err
	}
	a.modules[path] = resolved

	info := &ModuleInfo{}

	for i, mem := range resolved.Memories {
		info.Memories = append(info.Memories, MemoryInfo{
			Index:  i,
			Min:    mem.Min,
			Max:    mem.Max,
			HasMax: mem.HasMax,
		})
	}

	for i, tbl := range resolved.Tables {
		info.Tables = append(info.Tables, TableInfo{
			Index:  i,
			Min:    tbl.Limits.Min,
			Max:    tbl.Limits.Max,
			HasMax: tbl.Limits.HasMax,
		})
	}

	for i, glob := range resolved.Globals {
		info.Globals = append(info.Globals, GlobalInfo{
			Index:   i,
			Type:    glob.Type.Type.String(),
			Mutable: glob.Type.Mutable,
		})
	}

	for _, fn := range resolved.Functions {
		fi := FunctionInfo{
			Index:    fn.Index,
			Name:     fn.Name,
			Imported: fn.Imported,
		}
		if fn.Type != nil {
			fi.Params = len(fn.Type.Params)
			fi.Results = len(fn.Type.Results)
		}
		info.Functions = append(info.Functions, fi)
	}

	for _, exp := range resolved.Exports {
		kind := ""
		switch exp.Kind {
		case wasm.ExportFunc:
			kind = "func"
		case wasm.ExportMemory:
			kind = "memory"
		case wasm.ExportGlobal:
			kind = "global"
		case wasm.ExportTable:
			kind = "table"
		}
		info.Exports = append(info.Exports, ExportInfo{
			Name:  exp.Name,
			Kind:  kind,
			Index: exp.Index,
		})
	}

	return info, nil
}

func (a *App) DisassembleFunction(path string, index uint32) (string, error) {
	module := a.modules[path]
	if module == nil {
		return "", fmt.Errorf("module not loaded: %s", path)
	}
	fn := module.GetFunction(index)
	if fn == nil {
		return "", fmt.Errorf("function %d not found", index)
	}
	return formatFunctionWAT(fn, module), nil
}

func (a *App) DecompileFunction(path string, index uint32) (string, error) {
	module := a.modules[path]
	if module == nil {
		return "", fmt.Errorf("module not loaded: %s", path)
	}
	fn := module.GetFunction(index)
	if fn == nil {
		return "", fmt.Errorf("function %d not found", index)
	}
	if fn.Imported {
		return fmt.Sprintf("// imported: %s", fn.Name), nil
	}
	return decompile.Decompile(fn, module), nil
}

func (a *App) GetMemory(path string, index int) (string, error) {
	module := a.modules[path]
	if module == nil {
		return "", fmt.Errorf("module not loaded: %s", path)
	}
	if index < 0 || index >= len(module.Memories) {
		return "", fmt.Errorf("memory %d not found", index)
	}
	mem := module.Memories[index]
	if mem.HasMax {
		return fmt.Sprintf(";; Memory %d\n(memory %d %d)", index, mem.Min, mem.Max), nil
	}
	return fmt.Sprintf(";; Memory %d\n(memory %d)", index, mem.Min), nil
}

func (a *App) GetTable(path string, index int) (string, error) {
	module := a.modules[path]
	if module == nil {
		return "", fmt.Errorf("module not loaded: %s", path)
	}
	if index < 0 || index >= len(module.Tables) {
		return "", fmt.Errorf("table %d not found", index)
	}
	tbl := module.Tables[index]
	if tbl.Limits.HasMax {
		return fmt.Sprintf(";; Table %d\n(table %d %d funcref)", index, tbl.Limits.Min, tbl.Limits.Max), nil
	}
	return fmt.Sprintf(";; Table %d\n(table %d funcref)", index, tbl.Limits.Min), nil
}

func (a *App) GetGlobal(path string, index int) (string, error) {
	module := a.modules[path]
	if module == nil {
		return "", fmt.Errorf("module not loaded: %s", path)
	}
	if index < 0 || index >= len(module.Globals) {
		return "", fmt.Errorf("global %d not found", index)
	}
	glob := module.Globals[index]
	mut := ""
	if glob.Type.Mutable {
		mut = "(mut " + glob.Type.Type.String() + ")"
	} else {
		mut = glob.Type.Type.String()
	}
	var init string
	for _, instr := range glob.Init {
		if instr.Opcode == wasm.OpEnd {
			continue
		}
		init += fmt.Sprintf("(%s)", formatInstr(&instr))
	}
	return fmt.Sprintf(";; Global %d\n(global %s %s)", index, mut, init), nil
}

type MemoryData struct {
	Data       []byte        `json:"data"`
	TotalSize  int           `json:"totalSize"`
	Offset     int           `json:"offset"`
	Segments   []DataSegInfo `json:"segments"`
}

type DataSegInfo struct {
	Offset int `json:"offset"`
	Size   int `json:"size"`
}

type FunctionRef struct {
	Index uint32 `json:"index"`
	Name  string `json:"name"`
}

type XRefInfo struct {
	Callers []FunctionRef `json:"callers"`
	Callees []FunctionRef `json:"callees"`
}

func (a *App) GetMemoryData(path string, memIndex int, offset int, length int) (*MemoryData, error) {
	module := a.modules[path]
	if module == nil {
		return nil, fmt.Errorf("module not loaded: %s", path)
	}

	var segments []DataSegInfo
	for _, seg := range module.Data {
		segOffset := getDataSegOffset(seg.Offset)
		segments = append(segments, DataSegInfo{
			Offset: int(segOffset),
			Size:   len(seg.Data),
		})
	}

	mem := module.BuildMemory()
	if mem == nil {
		return &MemoryData{Data: []byte{}, TotalSize: 0, Offset: 0, Segments: segments}, nil
	}
	totalSize := len(mem)
	if offset >= totalSize {
		return &MemoryData{Data: []byte{}, TotalSize: totalSize, Offset: offset, Segments: segments}, nil
	}
	end := offset + length
	if end > totalSize {
		end = totalSize
	}
	return &MemoryData{
		Data:      mem[offset:end],
		TotalSize: totalSize,
		Offset:    offset,
		Segments:  segments,
	}, nil
}

func (a *App) GetXRefs(path string, funcIndex uint32) (*XRefInfo, error) {
	module := a.modules[path]
	if module == nil {
		return nil, fmt.Errorf("module not loaded: %s", path)
	}

	cg := decompile.BuildCallGraph(module)
	info := &XRefInfo{
		Callers: []FunctionRef{},
		Callees: []FunctionRef{},
	}

	for _, callerIdx := range cg.Callers[funcIndex] {
		fn := module.GetFunction(callerIdx)
		name := ""
		if fn != nil {
			name = fn.Name
		}
		if name == "" {
			name = fmt.Sprintf("func_%d", callerIdx)
		}
		info.Callers = append(info.Callers, FunctionRef{Index: callerIdx, Name: name})
	}

	for _, calleeIdx := range cg.Callees[funcIndex] {
		fn := module.GetFunction(calleeIdx)
		name := ""
		if fn != nil {
			name = fn.Name
		}
		if name == "" {
			name = fmt.Sprintf("func_%d", calleeIdx)
		}
		info.Callees = append(info.Callees, FunctionRef{Index: calleeIdx, Name: name})
	}

	return info, nil
}

type ErrorInfo struct {
	Offset  uint64 `json:"offset"`
	Opcode  string `json:"opcode"`
	Message string `json:"message"`
}

type FunctionErrorInfo struct {
	FuncIndex uint32      `json:"funcIndex"`
	FuncName  string      `json:"funcName"`
	Errors    []ErrorInfo `json:"errors"`
}

type ModuleErrorsInfo struct {
	Functions    []FunctionErrorInfo `json:"functions"`
	TotalErrors  int                 `json:"totalErrors"`
	UniqueErrors map[string]int      `json:"uniqueErrors"`
}

func (a *App) GetModuleErrors(path string) (*ModuleErrorsInfo, error) {
	module := a.modules[path]
	if module == nil {
		return nil, fmt.Errorf("module not loaded: %s", path)
	}

	errors := decompile.CollectErrors(module)
	result := &ModuleErrorsInfo{
		TotalErrors:  errors.TotalErrors,
		UniqueErrors: errors.UniqueErrors,
	}

	for _, fe := range errors.Functions {
		funcErr := FunctionErrorInfo{
			FuncIndex: fe.FuncIndex,
			FuncName:  fe.FuncName,
		}
		if funcErr.FuncName == "" {
			funcErr.FuncName = fmt.Sprintf("func_%d", fe.FuncIndex)
		}
		for _, e := range fe.Errors {
			funcErr.Errors = append(funcErr.Errors, ErrorInfo{
				Offset:  e.Offset,
				Opcode:  e.Opcode,
				Message: e.Message,
			})
		}
		result.Functions = append(result.Functions, funcErr)
	}

	return result, nil
}

func getDataSegOffset(instrs []wasm.Instruction) uint32 {
	for _, instr := range instrs {
		if instr.Opcode == wasm.OpI32Const && len(instr.Immediates) > 0 {
			if v, ok := instr.Immediates[0].(int32); ok {
				return uint32(v)
			}
			if v, ok := instr.Immediates[0].(uint32); ok {
				return v
			}
		}
	}
	return 0
}

func formatFunctionWAT(fn *wasm.ResolvedFunction, rm *wasm.ResolvedModule) string {
	if fn.Imported {
		for _, imp := range rm.Imports {
			if imp.Kind == wasm.ImportFunc {
				return fmt.Sprintf("(import %q %q (func))", imp.Module, imp.Name)
			}
		}
		return "(import)"
	}

	var result string
	result = fmt.Sprintf(";; Function %d: %s\n", fn.Index, fn.Name)
	result += "(func"

	if fn.Type != nil {
		if len(fn.Type.Params) > 0 {
			result += " (param"
			for _, p := range fn.Type.Params {
				result += " " + p.String()
			}
			result += ")"
		}
		if len(fn.Type.Results) > 0 {
			result += " (result"
			for _, r := range fn.Type.Results {
				result += " " + r.String()
			}
			result += ")"
		}
	}
	result += "\n"

	if fn.Body != nil {
		for _, loc := range fn.Body.Locals {
			for i := uint32(0); i < loc.Count; i++ {
				result += fmt.Sprintf("  (local %s)\n", wasm.ValType(loc.Type).String())
			}
		}
		for _, instr := range fn.Body.Instructions {
			if instr.Opcode == wasm.OpEnd {
				continue
			}
			result += "  " + formatInstr(&instr) + "\n"
		}
	}

	result += ")"
	return result
}

func formatInstr(instr *wasm.Instruction) string {
	if len(instr.Immediates) == 0 {
		return instr.Name
	}
	result := instr.Name
	for _, imm := range instr.Immediates {
		result += fmt.Sprintf(" %v", imm)
	}
	return result
}
