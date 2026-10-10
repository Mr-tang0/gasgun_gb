package main

import (
	"context"
	"fmt"
	"gasgun_gb/services"
	xinjiegasgun1 "gasgun_gb/services/xinjie_gasgun1"
	xinjiegasgun2 "gasgun_gb/services/xinjie_gasgun2"
	hepsgasgun1 "gasgun_gb/services/xinjie_heps_gasgun1"
)

type App struct {
	ctx           context.Context
	xinjieGasgun1 *xinjiegasgun1.XinjieGasGun1
	xinjieGasgun2 *xinjiegasgun2.XinjieGasGun2
	hepsGasgun1   *hepsgasgun1.HEPSGasGun1
	// normalHopkinson *normalhopkinson.NormalHopkinsonContoller

	user    string
	updater *services.UpdateService
}

func NewApp(user string,
	xinjieGasgun1 *xinjiegasgun1.XinjieGasGun1,
	xinjieGasgun2 *xinjiegasgun2.XinjieGasGun2,
	HEPSGasgun1 *hepsgasgun1.HEPSGasGun1) *App {
	app := &App{
		xinjieGasgun1: xinjieGasgun1,
		xinjieGasgun2: xinjieGasgun2,
		hepsGasgun1:   HEPSGasgun1,
		user:          user,
		updater:       services.NewUpdateService(),
	}
	return app
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	switch a.user {
	case "HEPS", "heps":
		a.hepsGasgun1.Startup(ctx)
		a.xinjieGasgun2.Startup(ctx)
	case "NIMTE", "nimte":
		a.xinjieGasgun1.Startup(ctx)
	case "PIMS", "pims":
		a.xinjieGasgun1.Startup(ctx)
		a.xinjieGasgun2.Startup(ctx)
	case "SWJTU", "swjtu":
		a.xinjieGasgun1.Startup(ctx)
	default:
		println("user not found")
	}

	a.APIUpdate()
}

func (a *App) GetUser() string {
	return a.user
}

func (a *App) APIUpdate() services.GitHubRelease {
	//获取更新信息
	release, err := a.updater.GetUpdateInfo()
	if err != nil {
		fmt.Printf("获取更新信息失败: %v\n", err)
		return services.GitHubRelease{}
	}
	fmt.Printf("更新信息: %v\n", release)
	return release
}

func (a *App) GetCachedRelease() *services.GitHubRelease {
	return a.updater.GetCachedRelease()
}
