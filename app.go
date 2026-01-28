package main

import (
	"context"
	"fmt"
	"gasgun_gb/backend"
)

type App struct {
	ctx             context.Context
	gasgun1         *backend.GasGun1Controller
	normalHopkinson *backend.NormalHopkinsonContoller
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) CallGasgun1() {
	fmt.Println("Call Gasgun1")
	a.gasgun1.Init(a.ctx)
}

func (a *App) CallNormalHopkinson() {
	fmt.Println("Call NormalHopkinson")
	a.normalHopkinson.Init(a.ctx)
}
