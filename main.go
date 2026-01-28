package main

import (
	"embed"
	"gasgun_gb/backend"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure

	gasgun1 := backend.NewGasGun1Controller()
	normalHopkinson := backend.NewNormalHopkinsonContoller()

	app := NewApp()
	app.gasgun1 = gasgun1
	app.normalHopkinson = normalHopkinson

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "力学加载实验室一站式操作平台",
		Width:  1200,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
			gasgun1,
			normalHopkinson,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
