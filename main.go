package main

import (
	"embed"
	"log"

	"github.com/yakutozcan/fast-ip-changer/pkg/sysexec"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// frontend/dist holds the Vite build output, which is not committed. The
// directory is kept non-empty by a tracked .gitkeep (restored on every build
// from frontend/public), because an embed pattern that matches no files is a
// compile error and would break `go build` on a fresh clone.
//
//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Changing the network configuration needs administrator rights. On Windows
	// the manifest asks for them at launch; on macOS each change is authorised
	// through a system prompt. Warn either way so a failure later is no surprise.
	if !sysexec.IsElevated() {
		log.Println("warning: running without administrator rights; network changes will ask for elevation")
	}

	// Create application with options
	err := wails.Run(&options.App{
		Title:         "Fast IP Changer",
		Width:         480,
		Height:        780,
		DisableResize: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Println("Error:", err.Error())
	}
}
