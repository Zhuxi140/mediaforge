package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"mediaforge/controller"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	renamerApp := controller.NewRenamerApp()
	mediaApp := controller.NewMediaApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "mediaforge",
		Width:  1312,
		Height: 720,

		MinWidth:  1000,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		OnStartup: func(ctx context.Context) {
			renamerApp.Startup(ctx)
			mediaApp.Startup(ctx)
		},
		Bind: []interface{}{
			renamerApp,
			mediaApp,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
