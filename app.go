package main

import (
	"context"
	"fmt"
	"gasgun_gb/backend"
	xinjiegasgun1 "gasgun_gb/services/xinjie_gasgun1"
	xinjiegasgun2 "gasgun_gb/services/xinjie_gasgun2"
)

type App struct {
	ctx             context.Context
	xinjieGasgun1   *xinjiegasgun1.XinjieGasGun1
	xinjieGasgun2   *xinjiegasgun2.XinjieGasGun2
	normalHopkinson *backend.NormalHopkinsonContoller

	updater *backend.UpdateService
}

func NewApp(xinjieGasgun1 *xinjiegasgun1.XinjieGasGun1, xinjieGasgun2 *xinjiegasgun2.XinjieGasGun2) *App {
	app := &App{
		ctx:           context.Background(),
		xinjieGasgun1: xinjieGasgun1,
		xinjieGasgun2: xinjieGasgun2,
	}
	app.xinjieGasgun1.Startup(app.ctx)
	app.xinjieGasgun2.Startup(app.ctx)
	return app
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	//创建更新服务
	a.updater = &backend.UpdateService{}
}

func (a *App) APIUpdate() backend.GitHubRelease {
	//获取更新信息
	release, err := a.updater.GetUpdateInfo()
	if err != nil {
		fmt.Printf("获取更新信息失败: %v\n", err)
		return backend.GitHubRelease{}
	}
	fmt.Printf("更新信息: %v\n", release)
	return release
}

func (a *App) GetCachedRelease() backend.GitHubRelease {
	return a.updater.GetCachedRelease()
}
