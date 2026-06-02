package backend

import (
	"context"
	"encoding/json"
	"fmt"
	protocol "gasgun_gb/backend/Protocol"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type GasGun1Controller struct {
	Siemens    *protocol.Siemens
	ctx        context.Context
	stopListen context.CancelFunc     // 用于停止 100ms 的轮询
	config     map[string]interface{} // 存储配置
}

func NewGasGun1Controller() *GasGun1Controller {
	return &GasGun1Controller{}
}

func (g *GasGun1Controller) Init(ctx context.Context) {
	g.ctx = ctx
	g.Siemens = protocol.NewSiemens(ctx)
	// 初始化时加载一次配置到内存
	g.GetConfig()
}

// 获取配置文件路径
func (g *GasGun1Controller) GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "config_gasgun1.json"
	}
	configDir := filepath.Join(homeDir, "Tang", "GASGUN_GB", "GasGun1")
	_ = os.MkdirAll(configDir, 0755)
	return filepath.Join(configDir, "config.json")
}

// 读取配置
func (g *GasGun1Controller) GetConfig() map[string]interface{} {
	configPath := g.GetConfigPath()
	data, err := os.ReadFile(configPath)

	// 默认配置（针对 GasGun1 的 5 路 VD 和相应 Q 点）
	defaultConfig := map[string]interface{}{
		"inletAddr":          "Q0.0",
		"outletAddr":         "Q0.1",
		"fireAddr":           "Q0.2",
		"vacuumAddr":         "Q0.3",
		"tailVacuumAddr":     "Q0.4",
		"tankPressureAddr":   "VD0",
		"supplyPressureAddr": "VD4",
		"targetVacuumAddr":   "VD8",
		"tailVacuumValAddr":  "VD12",
		"tankRange":          16.0,
		"supplyRange":        16.0,
		"ip":                 "192.168.2.1",
	}

	if err != nil {
		g.config = defaultConfig
		return g.config
	}

	var fileConfig map[string]interface{}
	if err := json.Unmarshal(data, &fileConfig); err != nil {
		g.config = defaultConfig
	} else {
		g.config = fileConfig
	}

	return g.config
}

// 保存配置
func (g *GasGun1Controller) SaveConfig(newConfig map[string]interface{}) APIResponse {
	g.config = newConfig
	configPath := g.GetConfigPath()

	data, _ := json.MarshalIndent(newConfig, "", "  ")
	err := os.WriteFile(configPath, data, 0644)
	if err != nil {
		return APIResponse{Status: false, Message: "保存配置失败！"}
	}
	return APIResponse{Status: true, Message: "配置保存成功！"}
}

// 连接PLC
func (g *GasGun1Controller) ConnectPLC(address string) APIResponse {
	fmt.Println("GasGun1 connecting to:", address)

	err := g.Siemens.Connect(address)
	if err != nil {
		return APIResponse{Status: false, Message: "连接失败！"}
	}

	ctx, cancel := context.WithCancel(context.Background())
	g.stopListen = cancel

	go func(ctx context.Context) {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				g.GetRealTimeData()
			case <-ctx.Done():
				fmt.Println("GasGun1 数据采集已停止")
				return
			}
		}
	}(ctx)

	return APIResponse{Status: true, Message: "连接成功，采集启动！"}
}

func (g *GasGun1Controller) DisconnectPLC() APIResponse {
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

func (g *GasGun1Controller) GetRealTimeData() {
	// 使用配置中的起始地址读取 5 个浮点数
	// 假设从 tankPressureAddr 开始连续读取
	startAddr := g.config["tankPressureAddr"].(string)
	data, err := g.Siemens.ReadVDFloats(startAddr, 5)

	if err != nil {
		fmt.Printf("[%s] GasGun1 读取失败: %v\n", time.Now().Format("15:04:05"), err)
		return
	}

	// 映射数据到前端
	result := map[string]float32{
		"气瓶压力":  data[0],
		"供气压力":  data[1],
		"靶室真空度": data[2],
		"尾部真空度": data[3],
	}

	runtime.EventsEmit(g.ctx, "update_metrics", result)
}

// 进气
func (g *GasGun1Controller) InletSwitch(open bool) APIResponse {
	addr := g.config["inletAddr"].(string)
	err := g.Siemens.WriteQ(addr, open)
	if err == nil {
		msg := "停止进气成功！"
		if open {
			msg = "开始进气成功！"
		}
		return APIResponse{Status: true, Message: msg}
	}
	return APIResponse{Status: false, Message: "操作失败！"}
}

// 排气
func (g *GasGun1Controller) ExhaustSwitch(open bool) APIResponse {
	addr := g.config["outletAddr"].(string)
	err := g.Siemens.WriteQ(addr, open)
	if err == nil {
		msg := "停止排气成功！"
		if open {
			msg = "开始排气成功！"
		}
		return APIResponse{Status: true, Message: msg}
	}
	return APIResponse{Status: false, Message: "操作失败！"}
}

// 发射逻辑
func (g *GasGun1Controller) FireSwitch(s int) APIResponse {
	addr := g.config["fireAddr"].(string)
	err := g.Siemens.WriteQ(addr, true)
	if err != nil {
		return APIResponse{Status: false, Message: "发射失败：线圈无法开启"}
	}

	go func(delayMs int, targetAddr string) {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
		err := g.Siemens.WriteQ(targetAddr, false)
		if err != nil {
			fmt.Printf("[报警] 发射后自动关闭失败: %v\n", err)
		}
	}(s, addr)

	return APIResponse{
		Status:  true,
		Message: fmt.Sprintf("发射指令已执行，脉冲宽度 %d ms", s),
	}
}

// 抽靶室真空
func (g *GasGun1Controller) VacuumSwitch(open bool) APIResponse {
	addr := g.config["vacuumAddr"].(string)
	err := g.Siemens.WriteQ(addr, open)
	if err == nil {
		msg := "停止抽靶室真空成功！"
		if open {
			msg = "开始抽靶室真空成功！"
		}
		return APIResponse{Status: true, Message: msg}
	}
	return APIResponse{Status: false, Message: "操作失败！"}
}

// 抽尾部真空
func (g *GasGun1Controller) TailVacuumSwitch(open bool) APIResponse {
	addr := g.config["tailVacuumAddr"].(string)
	err := g.Siemens.WriteQ(addr, open)
	if err == nil {
		msg := "停止抽尾部真空成功！"
		if open {
			msg = "开始抽尾部真空成功！"
		}
		return APIResponse{Status: true, Message: msg}
	}
	return APIResponse{Status: false, Message: "操作失败！"}
}

// 自动增压
func (g *GasGun1Controller) AutoPressurize(target float32) APIResponse {
	// 占位逻辑
	return APIResponse{Status: true, Message: "自动增压指令已发送！"}
}
