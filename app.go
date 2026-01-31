package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gasgun_gb/backend"
	"io"
	"net/http"
	"time"
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

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func (a *App) GetLatestRelease() (map[string]string, error) {
	repo := "Mr-tang0/GASGUN_GB"
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("HTTP请求失败: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Printf("GitHub 返回状态码: %d\n", resp.StatusCode)
	if resp.StatusCode != 200 {
		return nil, errors.New("获取最新发布信息失败")
	}

	bodyBytes, _ := io.ReadAll(resp.Body)

	var release GitHubRelease
	if err := json.Unmarshal(bodyBytes, &release); err != nil {
		fmt.Printf("JSON 解析失败: %v\n", err)
		return nil, err
	}

	downloadUrl := release.HTMLURL // 兜底跳转到发布页
	if len(release.Assets) > 0 {
		downloadUrl = release.Assets[0].BrowserDownloadURL
	}

	println("release", release.TagName, downloadUrl)

	return map[string]string{
		"version":     release.TagName,
		"downloadUrl": downloadUrl,
	}, nil
}

func (a *App) CallGasgun1() {
	fmt.Println("Call Gasgun1")
	a.gasgun1.Init(a.ctx)
}

func (a *App) CallNormalHopkinson() {
	fmt.Println("Call NormalHopkinson")
	a.normalHopkinson.Init(a.ctx)
}
