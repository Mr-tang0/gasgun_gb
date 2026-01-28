package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type NormalHopkinsonContoller struct {
	Siemens    *Siemens
	ctx        context.Context
	stopListen context.CancelFunc     // 用于停止 100ms 的轮询
	config     map[string]interface{} // 存储配置
}

func NewNormalHopkinsonContoller() *NormalHopkinsonContoller {
	return &NormalHopkinsonContoller{}
}

func (g *NormalHopkinsonContoller) Init(ctx context.Context) {
	g.ctx = ctx
	g.Siemens = NewSiemens(ctx)
}

func (g *NormalHopkinsonContoller) GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// 如果获取失败，降级使用当前目录
		return "config.json"
	}
	configDir := filepath.Join(homeDir, "Tang", "GASGUN_GB", "NormalHopkinson")

	_ = os.MkdirAll(configDir, 0755)
	return filepath.Join(configDir, "config.json")
}

func (g *NormalHopkinsonContoller) GetConfig() map[string]interface{} {
	configPath := g.GetConfigPath()
	data, err := os.ReadFile(configPath)

	// 默认配置
	defaultConfig := map[string]interface{}{
		"inletAddr":          "Q0.0",
		"outletAddr":         "Q0.1",
		"fireAddr":           "Q0.2",
		"tankPressureAddr":   "VD0",
		"supplyPressureAddr": "VD4",
		"tankRange":          16.0,
		"supplyRange":        16.0,
		"ip":                 "192.168.2.1",
	}

	if err != nil {
		g.config = defaultConfig
		return g.config
	}

	// 解析实时的 JSON 数据
	var fileConfig map[string]interface{}
	if err := json.Unmarshal(data, &fileConfig); err != nil {
		g.config = defaultConfig
	} else {
		g.config = fileConfig
	}

	return g.config
}

func (g *NormalHopkinsonContoller) SaveConfig(newConfig map[string]interface{}) APIResponse {
	fmt.Println("newConfig:", newConfig)

	g.config = newConfig
	configPath := g.GetConfigPath()

	data, _ := json.MarshalIndent(newConfig, "", "  ")
	err := os.WriteFile(configPath, data, 0644)
	fmt.Println("err:", err)
	if err != nil {
		return APIResponse{Status: false, Message: "保存配置失败！"}
	}
	return APIResponse{Status: true, Message: "配置保存成功！"}

}

// 连接PLC
func (g *NormalHopkinsonContoller) ConnectPLC(address string) APIResponse {
	fmt.Println("address:", address)

	err := g.Siemens.Connect(address)
	if err != nil {
		return APIResponse{Status: false, Message: "连接失败！"}
	}

	// 创建一个可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	g.stopListen = cancel

	// 开启 100ms 轮询协程
	go func(ctx context.Context) {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				g.GetRealTimeData()
			case <-ctx.Done():
				fmt.Println("数据采集协程已停止")
				return
			}
		}
	}(ctx)

	return APIResponse{Status: true, Message: "连接成功，采集启动！"}
}

func (g *NormalHopkinsonContoller) DisconnectPLC() APIResponse {
	fmt.Println("DisconnectPLC")

	if g.stopListen != nil {
		g.stopListen()
		g.stopListen = nil
	}

	err := g.Siemens.Disconnect()
	if err != nil {
		return APIResponse{Status: false, Message: "断开连接失败！"}
	}

	return APIResponse{Status: true, Message: "已断开连接并停止采集！"}
}

func (g *NormalHopkinsonContoller) GetRealTimeData() {

	data1, err := g.Siemens.ReadVDFloats(g.config["tankPressureAddr"].(string), 1)
	data2, err := g.Siemens.ReadVDFloats(g.config["supplyPressureAddr"].(string), 1)

	if err != nil {
		fmt.Printf("[%s] 读取失败: %v\n", time.Now().Format("15:04:05"), err)
	}

	result := map[string]float32{
		"气瓶压力": data1[0],
		"供气压力": data2[0],
	}

	runtime.EventsEmit(g.ctx, "normal_hopkinson_metrics", result)
}

// 进气
func (g *NormalHopkinsonContoller) InletSwitch(open bool) APIResponse {
	err := g.Siemens.WriteQ(g.config["inletAddr"].(string), open)
	if err == nil {
		if open {
			return APIResponse{Status: true, Message: "开始进气成功！"}
		} else {
			return APIResponse{Status: true, Message: "停止进气成功！"}
		}

	}
	return APIResponse{Status: false, Message: "操作失败！"}
}

//排气

func (g *NormalHopkinsonContoller) ExhaustSwitch(open bool) APIResponse {
	err := g.Siemens.WriteQ(g.config["outletAddr"].(string), open)
	if err == nil {
		if open {
			return APIResponse{Status: true, Message: "开始排气成功！"}
		} else {
			return APIResponse{Status: true, Message: "停止排气成功！"}
		}

	}
	return APIResponse{Status: false, Message: "操作失败！"}
}

// FireSwitch 发射逻辑：打开 Q0.2，持续 s 毫秒后自动关闭
func (g *NormalHopkinsonContoller) FireSwitch(s int) APIResponse {
	err := g.Siemens.WriteQ(g.config["fireAddr"].(string), true)
	if err != nil {
		return APIResponse{Status: false, Message: "发射失败：线圈无法开启"}
	}

	go func(delayMs int) {
		// 延时指定的毫秒数
		time.Sleep(time.Duration(delayMs) * time.Millisecond)

		err := g.Siemens.WriteQ(g.config["fireAddr"].(string), false)
		if err != nil {
			// 这里建议记录日志，因为协程内部的错误无法直接返回给前端 API
			fmt.Printf("[报警] 发射后自动关闭失败: %v\n", err)
		}
	}(s)

	return APIResponse{
		Status:  true,
		Message: fmt.Sprintf("发射指令已执行，脉冲宽度 %d ms", s),
	}
}

// 自动增压
func (g *NormalHopkinsonContoller) AutoPressurize(target float32) APIResponse {
	return APIResponse{Status: true, Message: "自动增压成功！"}
}
