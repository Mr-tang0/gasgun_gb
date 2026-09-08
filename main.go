package main

import (
	"embed"
	xinjiegasgun1 "gasgun_gb/services/xinjie_gasgun1"
	xinjiegasgun2 "gasgun_gb/services/xinjie_gasgun2"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure

	XinjieGasgun1 := xinjiegasgun1.NewXinjieGasGun1()
	XinjieGasgun2 := xinjiegasgun2.NewXinjieGasGun2()

	app := NewApp(XinjieGasgun1, XinjieGasgun2)

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "力学加载实验室一站式操作平台",
		Width:  1300,
		Height: 930,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Frameless:        true,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
			XinjieGasgun1,
			XinjieGasgun2,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
