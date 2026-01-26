package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	appMenu := menu.NewMenu()

	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("Open...", keys.CmdOrCtrl("o"), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:open")
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Save Annotations", keys.CmdOrCtrl("s"), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:save")
	})
	fileMenu.AddText("Import Annotations...", nil, func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:import")
	})
	fileMenu.AddText("Export Annotations...", nil, func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:export")
	})
	fileMenu.AddText("Clear Annotations", nil, func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:clear")
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Quit", keys.CmdOrCtrl("q"), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:quit")
	})

	editMenu := appMenu.AddSubmenu("Edit")
	editMenu.AddText("Copy", keys.CmdOrCtrl("c"), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:copy")
	})
	editMenu.AddText("Select All", keys.CmdOrCtrl("a"), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:selectall")
	})
	editMenu.AddSeparator()
	editMenu.AddText("Preferences...", keys.CmdOrCtrl(","), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:preferences")
	})

	viewMenu := appMenu.AddSubmenu("View")
	viewMenu.AddText("Toggle Fullscreen", keys.Key("F11"), func(cd *menu.CallbackData) {
		runtime.WindowToggleMaximise(app.ctx)
	})

	helpMenu := appMenu.AddSubmenu("Help")
	helpMenu.AddText("About", nil, func(cd *menu.CallbackData) {
		runtime.MessageDialog(app.ctx, runtime.MessageDialogOptions{
			Type:    runtime.InfoDialog,
			Title:   "About wasmspy",
			Message: "WebAssembly Disassembler & Decompiler",
		})
	})

	err := wails.Run(&options.App{
		Title:  "wasmspy",
		Width:  1920,
		Height: 1080,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true,
		},
		OnDomReady: app.onDomReady,
		Menu:       appMenu,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
