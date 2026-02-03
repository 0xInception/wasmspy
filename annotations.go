package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type FunctionAnnotation struct {
	Name    string `json:"name,omitempty"`
	Comment string `json:"comment,omitempty"`
}

type Annotations struct {
	Version           int                            `json:"version"`
	Functions         map[string]*FunctionAnnotation `json:"functions,omitempty"`
	Comments          map[string]string              `json:"comments,omitempty"`
	DecompileComments map[string]string              `json:"decompileComments,omitempty"`
	Bookmarks         []uint32                       `json:"bookmarks,omitempty"`
}

func NewAnnotations() *Annotations {
	return &Annotations{
		Version:           1,
		Functions:         make(map[string]*FunctionAnnotation),
		Comments:          make(map[string]string),
		DecompileComments: make(map[string]string),
		Bookmarks:         []uint32{},
	}
}

func annotationsPath(wasmPath string) string {
	return wasmPath + ".wasmspy"
}

func loadAnnotationsFromFile(wasmPath string) *Annotations {
	data, err := os.ReadFile(annotationsPath(wasmPath))
	if err != nil {
		return NewAnnotations()
	}
	var ann Annotations
	if err := json.Unmarshal(data, &ann); err != nil {
		return NewAnnotations()
	}
	if ann.Functions == nil {
		ann.Functions = make(map[string]*FunctionAnnotation)
	}
	if ann.Comments == nil {
		ann.Comments = make(map[string]string)
	}
	if ann.DecompileComments == nil {
		ann.DecompileComments = make(map[string]string)
	}
	return &ann
}

func saveAnnotationsToFile(wasmPath string, ann *Annotations) error {
	data, err := json.MarshalIndent(ann, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(annotationsPath(wasmPath), data, 0644)
}

func (a *App) GetAnnotations(path string) *Annotations {
	if ann := a.annotations[path]; ann != nil {
		return ann
	}
	return NewAnnotations()
}

func (a *App) SetFunctionName(path string, index uint32, name string) error {
	ann := a.annotations[path]
	if ann == nil {
		ann = NewAnnotations()
		a.annotations[path] = ann
	}
	key := fmt.Sprintf("%d", index)
	if ann.Functions[key] == nil {
		ann.Functions[key] = &FunctionAnnotation{}
	}
	ann.Functions[key].Name = name
	return nil
}

func (a *App) SetFunctionComment(path string, index uint32, comment string) error {
	ann := a.annotations[path]
	if ann == nil {
		ann = NewAnnotations()
		a.annotations[path] = ann
	}
	key := fmt.Sprintf("%d", index)
	if ann.Functions[key] == nil {
		ann.Functions[key] = &FunctionAnnotation{}
	}
	ann.Functions[key].Comment = comment
	return nil
}

func (a *App) SetOffsetComment(path string, offset uint64, comment string, isDecompile bool) error {
	ann := a.annotations[path]
	if ann == nil {
		ann = NewAnnotations()
		a.annotations[path] = ann
	}
	key := fmt.Sprintf("0x%x", offset)
	target := ann.Comments
	if isDecompile {
		target = ann.DecompileComments
	}
	if comment == "" {
		delete(target, key)
	} else {
		target[key] = comment
	}
	return nil
}

func (a *App) SetBookmarks(path string, bookmarks []uint32) error {
	ann := a.annotations[path]
	if ann == nil {
		ann = NewAnnotations()
		a.annotations[path] = ann
	}
	ann.Bookmarks = bookmarks
	return nil
}

func (a *App) SaveAnnotations(path string) error {
	ann := a.annotations[path]
	if ann == nil {
		return nil
	}
	return saveAnnotationsToFile(path, ann)
}

func (a *App) ClearAnnotations(path string) *Annotations {
	a.annotations[path] = NewAnnotations()
	return a.annotations[path]
}

func (a *App) ExportAnnotations(path string) (string, error) {
	ann := a.annotations[path]
	if ann == nil {
		ann = NewAnnotations()
	}
	data, err := json.MarshalIndent(ann, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (a *App) ImportAnnotations(path string, data string) error {
	var ann Annotations
	if err := json.Unmarshal([]byte(data), &ann); err != nil {
		return err
	}
	if ann.Functions == nil {
		ann.Functions = make(map[string]*FunctionAnnotation)
	}
	if ann.Comments == nil {
		ann.Comments = make(map[string]string)
	}
	if ann.DecompileComments == nil {
		ann.DecompileComments = make(map[string]string)
	}
	a.annotations[path] = &ann
	return nil
}

func (a *App) ExportAnnotationsToFile(wasmPath string) (string, error) {
	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export Annotations",
		DefaultFilename: "annotations.wasmspy",
		Filters: []runtime.FileFilter{
			{DisplayName: "WasmSpy Annotations", Pattern: "*.wasmspy"},
			{DisplayName: "JSON Files", Pattern: "*.json"},
		},
	})
	if err != nil || savePath == "" {
		return "", err
	}
	ann := a.annotations[wasmPath]
	if ann == nil {
		ann = NewAnnotations()
	}
	data, err := json.MarshalIndent(ann, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return "", err
	}
	return savePath, nil
}

func (a *App) ImportAnnotationsFromFile(wasmPath string) (string, error) {
	openPath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Import Annotations",
		Filters: []runtime.FileFilter{
			{DisplayName: "WasmSpy Annotations", Pattern: "*.wasmspy"},
			{DisplayName: "JSON Files", Pattern: "*.json"},
		},
	})
	if err != nil || openPath == "" {
		return "", err
	}
	data, err := os.ReadFile(openPath)
	if err != nil {
		return "", err
	}
	var ann Annotations
	if err := json.Unmarshal(data, &ann); err != nil {
		return "", err
	}
	if ann.Functions == nil {
		ann.Functions = make(map[string]*FunctionAnnotation)
	}
	if ann.Comments == nil {
		ann.Comments = make(map[string]string)
	}
	if ann.DecompileComments == nil {
		ann.DecompileComments = make(map[string]string)
	}
	a.annotations[wasmPath] = &ann
	return openPath, nil
}
