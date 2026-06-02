package main

import (
	"context"
	"fmt"
	"gasgun_gb/backend"
)

type App struct {
	ctx             context.Context
	gasgun1         *backend.GasGun1Controller
	gasgun2         *backend.GasGun2Controller
	normalHopkinson *backend.NormalHopkinsonContoller

	updater *backend.UpdateService
}

func NewApp() *App {
	return &App{}
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

func (a *App) CallGasgun1() {
	fmt.Println("Call Gasgun1")
	a.gasgun1.Init(a.ctx)
}

func (a *App) CallNormalHopkinson() {
	fmt.Println("Call NormalHopkinson")
	a.normalHopkinson.Init(a.ctx)
}

func (a *App) CallGasGun2() {
	fmt.Println("Call GasGun2")
	a.gasgun2.Init(a.ctx)
}
