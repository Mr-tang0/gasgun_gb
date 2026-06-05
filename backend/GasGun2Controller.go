package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	protocol "gasgun_gb/backend/Protocol"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type SWITCHAddress struct {
	Pressurize         uint16 `json:"Pressurize"`         //增压阀1
	Decompress         uint16 `json:"Decompress"`         //减压阀1
	PumpTubePressurize uint16 `json:"PumpTubePressurize"` //泵管增压阀1
	PumpTubeDecompress uint16 `json:"PumpTubeDecompress"` //泵管减压阀1
	PumpTubeVacuum     uint16 `json:"PumpTubeVacuum"`     //抽泵管真空阀1
	TargetVacuum       uint16 `json:"TargetVacuum"`       //靶室真空阀1
	TailVacuumProtect  uint16 `json:"TailVacuumProtect"`  //尾真空保护阀（尾部真空阀）
	PumpTubeProtect    uint16 `json:"PumpTubeProtect"`    //泵管保护阀1
	FireSwitch         uint16 `json:"FireSwitch"`         //发射阀1
	SystemDecompress   uint16 `json:"SystemDecompress"`   //系统减压阀1
	TargetVacuumPump   uint16 `json:"TargetVacuumPump"`   //靶室真空泵1
	TailVacuumPump     uint16 `json:"TailVacuumPump"`     //尾真空泵1
}

// DataAddress 数据地址配置
type DataAddress struct {
	InputPressure      uint16 `json:"InputPressure"`      // 输入压力地址
	CylinderPressure   uint16 `json:"CylinderPressure"`   // 气瓶压力地址
	PumpTubePressure   uint16 `json:"PumpTubePressure"`   // 泵管压力地址
	PumpTubePressureHi uint16 `json:"PumpTubePressureHi"` // 泵管压力高精度地址
	TargetVacuumDegree uint16 `json:"TargetVacuumDegree"` // 靶室真空度地址
	TailVacuumDegree   uint16 `json:"TailVacuumDegree"`   // 尾部真空度地址
}

// GasGun2Config 配置结构体
type GasGun2Config struct {
	IP            string        `json:"ip"`            // PLC IP地址
	Switches      SWITCHAddress `json:"switches"`      // 各阀门的Modbus点位地址
	DataAddresses DataAddress   `json:"dataAddresses"` // 监控数据地址配置
}

// 监控数据结构体
type GasGun2Metrics struct {
	InputPressure      float32 `json:"inputPressure"`      //输入压力
	CylinderPressure   float32 `json:"cylinderPressure"`   //气瓶压力（一级气室）
	PumpTubePressure   float32 `json:"pumpTubePressure"`   //泵管压力（二级气室）
	PumpTubePressureHi float32 `json:"pumpTubePressureHi"` //泵管压力（高精度）
	TargetVacuumDegree float32 `json:"targetVacuumDegree"` //靶室真空度
	TailVacuumDegree   float32 `json:"tailVacuumDegree"`   //尾部真空度
}

type GasGun2Controller struct {
	ctx        context.Context
	stopListen context.CancelFunc // 用于停止轮询
	config     GasGun2Config      // 配置结构体

	// PLC 连接
	xinjieClient *protocol.XinjieClient

	// 自动压力控制相关
	pumpTubeTargetPressure float32
	cylinderTargetPressure float32
	stopPumpTubeControl    context.CancelFunc
	stopCylinderControl    context.CancelFunc
	isExternalTrigger      bool // 是否为外触发模式
}

func NewGasGun2Controller() *GasGun2Controller {
	return &GasGun2Controller{
		xinjieClient: &protocol.XinjieClient{},
	}
}

func (g *GasGun2Controller) Init(ctx context.Context) {
	g.ctx = ctx
	fmt.Println("GasGun2Controller init", g.ctx == nil)
	// 初始化时加载一次配置到内存
	g.GetConfig()
}

// 获取配置文件路径
func (g *GasGun2Controller) GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "config_gasgun2.json"
	}
	configDir := filepath.Join(homeDir, "Tang", "GASGUN_GB", "GasGun2")
	_ = os.MkdirAll(configDir, 0755)
	return filepath.Join(configDir, "config.json")
}

// 读取配置
func (g *GasGun2Controller) GetConfig() GasGun2Config {
	configPath := g.GetConfigPath()
	fmt.Println("读取配置：", configPath)
	data, err := os.ReadFile(configPath)

	// 默认配置
	defaultConfig := GasGun2Config{
		IP: "192.168.6.6",
		Switches: SWITCHAddress{
			Pressurize:         0,  // 增压阀1
			Decompress:         1,  // 减压阀1
			PumpTubePressurize: 4,  // 泵管增压阀1
			PumpTubeDecompress: 5,  // 泵管减压阀1
			PumpTubeVacuum:     7,  // 抽泵管真空阀1
			TargetVacuum:       13, // 靶室真空阀1
			TailVacuumProtect:  10, // 尾真空保护阀
			PumpTubeProtect:    6,  // 泵管保护阀1
			FireSwitch:         2,  // 发射阀1
			SystemDecompress:   3,  // 系统减压阀1
			TargetVacuumPump:   11, // 靶室真空泵1
			TailVacuumPump:     12, // 尾真空泵1
		},
		DataAddresses: DataAddress{
			InputPressure:      56, // 输入压力地址
			CylinderPressure:   54, // 气瓶压力地址
			PumpTubePressure:   50, // 泵管压力地址
			PumpTubePressureHi: 52, // 泵管压力高精度地址
			TargetVacuumDegree: 58, // 靶室真空度地址
			TailVacuumDegree:   60, // 尾部真空度地址
		},
	}

	if err != nil {
		g.config = defaultConfig
		return g.config
	}

	var fileConfig GasGun2Config
	if err := json.Unmarshal(data, &fileConfig); err != nil {
		g.config = defaultConfig
	} else {
		g.config = fileConfig
	}

	return g.config
}

// 保存配置
func (g *GasGun2Controller) SaveConfig(newConfig GasGun2Config) APIResponse {
	g.config = newConfig
	configPath := g.GetConfigPath()
	fmt.Println("保存配置：", configPath)

	data, _ := json.MarshalIndent(newConfig, "", "  ")
	err := os.WriteFile(configPath, data, 0644)
	if err != nil {
		return APIResponse{Status: false, Message: "保存配置失败！"}
	}
	return APIResponse{Status: true, Message: "配置保存成功！"}
}

// 设置触发模式：外触发开14和15引脚，内触发则关闭14和15引脚，同时均关闭发射引脚
func (g *GasGun2Controller) SetTriggerMode(isExternal bool) {
	err := g.CloseSwitch("FireSwitch") // 无论内外触发都先关闭发射阀，确保安全
	if err != nil {
		fmt.Printf("关闭发射阀失败: %v\n", err)
		return
	}

	flag := g.xinjieClient.Write_Y_Coils(15, []bool{isExternal, isExternal})
	if !flag {
		fmt.Printf("设置触发模式失败: %v\n", flag)
		return
	}

	g.isExternalTrigger = isExternal
	fmt.Printf("触发模式已设置为: %s\n", map[bool]string{true: "外触发", false: "内触发"}[isExternal])
}

// *********************************** PLC通信区域 *********************************

func (g *GasGun2Controller) ConnectPLC(ip string) APIResponse {
	fmt.Println("GasGun2 connecting to:", ip)

	// 更新配置中的 IP
	g.config.IP = ip
	g.SaveConfig(g.config)

	// 创建真实的 XinjieClient 连接
	addr := ip + ":502" // 默认 Modbus TCP 端口
	err := g.xinjieClient.OpenTCP(addr, 1)
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		return APIResponse{Status: false, Message: fmt.Sprintf("连接失败: %v", err)}
	}

	// 启动数据采集
	ctx, cancel := context.WithCancel(context.Background())
	g.stopListen = cancel

	go func(ctx context.Context) {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				g.GetRealTimeData()
			case <-ctx.Done():
				fmt.Println("GasGun2 数据采集已停止")
				return
			}
		}
	}(ctx)

	return APIResponse{Status: true, Message: "连接成功，采集启动！"}
}

func (g *GasGun2Controller) DisconnectPLC() APIResponse {
	if g.stopListen != nil {
		g.stopListen()
		g.stopListen = nil
	}

	// 停止自动压力控制
	if g.stopPumpTubeControl != nil {
		g.stopPumpTubeControl()
		g.stopPumpTubeControl = nil
	}
	if g.stopCylinderControl != nil {
		g.stopCylinderControl()
		g.stopCylinderControl = nil
	}

	// 关闭真实连接
	if g.xinjieClient != nil {
		g.xinjieClient.Close()
	}

	return APIResponse{Status: true, Message: "已断开连接并停止采集！"}
}

func (g *GasGun2Controller) GetRealTimeData() {
	if !g.xinjieClient.IsOpened() {
		fmt.Println("PLC client not connected")
		return
	}

	g.xinjieClient.Write_M_Coils(0, []bool{true})

	// 从配置中获取数据地址
	addr := g.config.DataAddresses

	// 每个浮点数占2个寄存器，从输入压力地址开始连续读取6个数据
	values := g.xinjieClient.ReadFloat32(50, 6)
	if len(values) < 6 {
		fmt.Println("Failed to read metrics from PLC")
		return
	}

	metrics := GasGun2Metrics{
		InputPressure:      values[(addr.InputPressure-50)/2],      // 输入压力
		CylinderPressure:   values[(addr.CylinderPressure-50)/2],   // 气瓶压力（一级气室）
		PumpTubePressure:   values[(addr.PumpTubePressure-50)/2],   // 泵管压力（二级气室）
		PumpTubePressureHi: values[(addr.PumpTubePressureHi-50)/2], // 泵管压力（高精度）
		TargetVacuumDegree: values[(addr.TargetVacuumDegree-50)/2], // 靶室真空度
		TailVacuumDegree:   values[(addr.TailVacuumDegree-50)/2],   // 尾部真空度
	}

	// 发送数据到前端
	if g.ctx == nil {
		fmt.Println("ctx is nil")
		return
	}

	runtime.EventsEmit(g.ctx, "update_gasgun2_metrics", metrics)
}

// *********************************** 控制阀门区域 *********************************

// 根据阀门名称获取Modbus点位地址
func (g *GasGun2Controller) getSwitchAddress(switchName string) (uint16, error) {
	switch switchName {
	case "Pressurize":
		return g.config.Switches.Pressurize, nil
	case "Decompress":
		return g.config.Switches.Decompress, nil
	case "PumpTubePressurize":
		return g.config.Switches.PumpTubePressurize, nil
	case "PumpTubeDecompress":
		return g.config.Switches.PumpTubeDecompress, nil
	case "PumpTubeVacuum":
		return g.config.Switches.PumpTubeVacuum, nil
	case "TargetVacuum":
		return g.config.Switches.TargetVacuum, nil
	case "TailVacuumProtect":
		return g.config.Switches.TailVacuumProtect, nil
	case "PumpTubeProtect":
		return g.config.Switches.PumpTubeProtect, nil
	case "FireSwitch":
		return g.config.Switches.FireSwitch, nil
	case "SystemDecompress":
		return g.config.Switches.SystemDecompress, nil
	case "TargetVacuumPump":
		return g.config.Switches.TargetVacuumPump, nil
	case "TailVacuumPump":
		return g.config.Switches.TailVacuumPump, nil
	default:
		return 0, fmt.Errorf("unknown switch: %s", switchName)
	}
}

func (g *GasGun2Controller) OpenSwitch(switchName string) error {
	// 检查PLC连接状态
	if !g.xinjieClient.IsOpened() {
		return fmt.Errorf("PLC未连接，请先连接设备")
	}

	address, err := g.getSwitchAddress(switchName)
	if err != nil {
		return err
	}

	flag := g.xinjieClient.Write_Y_Coils(address, []bool{true})
	fmt.Printf("打开阀门: %s (Modbus address: %d), result: %v\n", switchName, address, flag)

	if !flag {
		return fmt.Errorf("打开%s失败", switchName)
	}

	return nil
}

func (g *GasGun2Controller) CloseSwitch(switchName string) error {
	// 检查PLC连接状态
	if !g.xinjieClient.IsOpened() {
		return fmt.Errorf("PLC未连接，请先连接设备")
	}

	address, err := g.getSwitchAddress(switchName)
	if err != nil {
		return err
	}
	flag := g.xinjieClient.Write_Y_Coils(address, []bool{false})
	fmt.Printf("关闭阀门: %s (Modbus address: %d), result: %v\n", switchName, address, flag)

	if !flag {
		return fmt.Errorf("关闭%s失败", switchName)
	}
	return nil
}

//****************************************自动化流程*****************************************

// 1. 抽真空函数：开尾部真空泵-间隔1s-开尾部真空保护阀-间隔1s-开靶室真空泵
func (g *GasGun2Controller) StartAutoVacuum() APIResponse {
	// 开尾部真空泵
	err := g.OpenSwitch("TailVacuumPump")
	if err != nil {
		return APIResponse{Status: false, Message: "打开尾部真空泵失败"}
	}

	time.Sleep(1 * time.Second)

	// 开尾部真空保护阀
	err = g.OpenSwitch("TailVacuumProtect")
	if err != nil {
		g.CloseSwitch("TailVacuumPump")
		return APIResponse{Status: false, Message: "打开尾部真空保护阀失败"}
	}

	time.Sleep(1 * time.Second)

	// 开靶室真空泵
	err = g.OpenSwitch("TargetVacuumPump")
	if err != nil {
		g.CloseSwitch("TailVacuumPump")
		g.CloseSwitch("TailVacuumProtect")
		return APIResponse{Status: false, Message: "打开靶室真空泵失败"}
	}

	return APIResponse{Status: true, Message: "抽真空流程已启动"}
}

// 停止抽真空
func (g *GasGun2Controller) StopAutoVacuum() APIResponse {
	err := g.CloseSwitch("TailVacuumPump")
	if err != nil {
		return APIResponse{Status: false, Message: "关闭尾部真空泵失败"}
	}

	time.Sleep(1 * time.Second)

	err = g.CloseSwitch("TailVacuumProtect")
	if err != nil {
		return APIResponse{Status: false, Message: "关闭尾部真空保护阀失败"}
	}

	time.Sleep(1 * time.Second)

	err = g.CloseSwitch("TargetVacuumPump")
	if err != nil {
		return APIResponse{Status: false, Message: "关闭靶室真空泵失败"}
	}

	return APIResponse{Status: true, Message: "抽真空流程已停止"}
}

// 2. 抽泵管函数：打开抽泵管真空阀
func (g *GasGun2Controller) StartPumpTubeVacuum() APIResponse {
	err := g.OpenSwitch("PumpTubeVacuum")
	if err != nil {
		return APIResponse{Status: false, Message: "打开抽泵管真空阀失败"}
	}
	return APIResponse{Status: true, Message: "抽泵管真空已启动"}
}

func (g *GasGun2Controller) StopPumpTubeVacuum() APIResponse {
	err := g.CloseSwitch("PumpTubeVacuum")
	if err != nil {
		return APIResponse{Status: false, Message: "关闭抽泵管真空阀失败"}
	}
	return APIResponse{Status: true, Message: "抽泵管真空已停止"}
}

// 3. 子线程自动设置泵管压力（通过开关泵管增压阀与泵管减压阀）
func (g *GasGun2Controller) AutoPumpTubePressure(target float32) APIResponse {
	g.pumpTubeTargetPressure = target

	// 如果已有控制线程运行，先停止
	if g.stopPumpTubeControl != nil {
		g.stopPumpTubeControl()
	}

	ctx, cancel := context.WithCancel(context.Background())
	g.stopPumpTubeControl = cancel

	go func(ctx context.Context, target float32) {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 获取当前泵管压力
				currentPressure := g.getPumpTubePressure()

				// 控制逻辑：根据当前压力与目标压力的差值来控制阀门
				if math.Abs(float64(currentPressure-target)) > 0.1 {
					if currentPressure < target {
						// 需要增压：打开增压阀，关闭减压阀
						g.OpenSwitch("PumpTubePressurize")
						g.CloseSwitch("PumpTubeDecompress")
					} else {
						// 需要减压：打开减压阀，关闭增压阀
						g.OpenSwitch("PumpTubeDecompress")
						g.CloseSwitch("PumpTubePressurize")
					}
				} else {
					// 达到目标压力：关闭两个阀门
					g.CloseSwitch("PumpTubePressurize")
					g.CloseSwitch("PumpTubeDecompress")
					fmt.Printf("泵管压力已达到目标值: %.2f\n", target)
					return
				}
			case <-ctx.Done():
				// 停止时关闭阀门
				g.CloseSwitch("PumpTubePressurize")
				g.CloseSwitch("PumpTubeDecompress")
				fmt.Println("泵管自动压力控制已停止")
				return
			}
		}
	}(ctx, target)

	return APIResponse{Status: true, Message: fmt.Sprintf("泵管自动压力控制已启动，目标压力: %.2f", target)}
}

func (g *GasGun2Controller) StopAutoPumpTubePressure() APIResponse {
	if g.stopPumpTubeControl != nil {
		g.stopPumpTubeControl()
		g.stopPumpTubeControl = nil
	}
	return APIResponse{Status: true, Message: "泵管自动压力控制已停止"}
}

// 4. 子线程自动设置气瓶压力（通过开关增压阀与减压阀）
func (g *GasGun2Controller) AutoCylinderPressure(target float32) APIResponse {
	g.cylinderTargetPressure = target

	// 如果已有控制线程运行，先停止
	if g.stopCylinderControl != nil {
		g.stopCylinderControl()
	}

	ctx, cancel := context.WithCancel(context.Background())
	g.stopCylinderControl = cancel

	go func(ctx context.Context, target float32) {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 获取当前气瓶压力
				currentPressure := g.getCylinderPressure()

				// 控制逻辑
				if math.Abs(float64(currentPressure-target)) > 0.1 {
					if currentPressure < target {
						// 需要增压
						g.OpenSwitch("Pressurize")
						g.CloseSwitch("Decompress")
					} else {
						// 需要减压
						g.OpenSwitch("Decompress")
						g.CloseSwitch("Pressurize")
					}
				} else {
					// 达到目标压力
					g.CloseSwitch("Pressurize")
					g.CloseSwitch("Decompress")
					fmt.Printf("气瓶压力已达到目标值: %.2f\n", target)
					return
				}
			case <-ctx.Done():
				g.CloseSwitch("Pressurize")
				g.CloseSwitch("Decompress")
				fmt.Println("气瓶自动压力控制已停止")
				return
			}
		}
	}(ctx, target)

	return APIResponse{Status: true, Message: fmt.Sprintf("气瓶自动压力控制已启动，目标压力: %.2f", target)}
}

func (g *GasGun2Controller) StopAutoCylinderPressure() APIResponse {
	if g.stopCylinderControl != nil {
		g.stopCylinderControl()
		g.stopCylinderControl = nil
	}
	return APIResponse{Status: true, Message: "气瓶自动压力控制已停止"}
}

// 5. 准备发射：关闭尾部真空泵，尾部真空阀，靶室真空泵，打开发射阀（内触发时）
func (g *GasGun2Controller) PrepareFire() APIResponse {
	// 关闭尾部真空泵
	err := g.CloseSwitch("TailVacuumPump")
	if err != nil {
		return APIResponse{Status: false, Message: "关闭尾部真空泵失败"}
	}

	time.Sleep(500 * time.Millisecond)

	// 关闭尾部真空阀
	err = g.CloseSwitch("TailVacuumProtect")
	if err != nil {
		return APIResponse{Status: false, Message: "关闭尾部真空阀失败"}
	}

	time.Sleep(500 * time.Millisecond)

	// 关闭靶室真空泵
	err = g.CloseSwitch("TargetVacuumPump")
	if err != nil {
		return APIResponse{Status: false, Message: "关闭靶室真空泵失败"}
	}

	time.Sleep(500 * time.Millisecond)

	// 判断是内触发还是外触发，外触发时不打开发射阀
	if !g.isExternalTrigger {
		// 内触发：打开发射阀
		err = g.OpenSwitch("FireSwitch")
		if err != nil {
			return APIResponse{Status: false, Message: "打开发射阀失败"}
		}
		return APIResponse{Status: true, Message: "准备发射完成（内触发模式，发射阀已打开）"}
	} else {
		return APIResponse{Status: true, Message: "准备发射完成（外触发模式，等待外部触发信号）"}
	}
}

// 6. 恢复：打开减压阀、泵管减压阀、重新开启靶室真空泵
func (g *GasGun2Controller) ResetSystem(reset bool) APIResponse {
	if reset {
		// 打开减压阀
		err := g.OpenSwitch("Decompress")
		if err != nil {
			return APIResponse{Status: false, Message: "打开减压阀失败"}
		}

		time.Sleep(1000 * time.Millisecond)

		// 打开泵管减压阀
		err = g.OpenSwitch("PumpTubeDecompress")
		if err != nil {
			return APIResponse{Status: false, Message: "打开泵管减压阀失败"}
		}

		time.Sleep(1000 * time.Millisecond)

		// 重新开启靶室真空泵
		err = g.OpenSwitch("SystemDecompress")
		if err != nil {
			return APIResponse{Status: false, Message: "开启靶室真空泵失败"}
		}

		return APIResponse{Status: true, Message: "系统恢复完成"}
	} else {
		// 关闭减压阀
		err := g.CloseSwitch("Decompress")
		if err != nil {
			return APIResponse{Status: false, Message: "关闭减压阀失败"}
		}

		time.Sleep(1000 * time.Millisecond)

		// 关闭泵管减压阀
		err = g.CloseSwitch("PumpTubeDecompress")
		if err != nil {
			return APIResponse{Status: false, Message: "关闭泵管减压阀失败"}
		}

		time.Sleep(1000 * time.Millisecond)

		err = g.CloseSwitch("SystemDecompress")
		if err != nil {
			return APIResponse{Status: false, Message: "关闭目标真空泵阀失败"}
		}

		return APIResponse{Status: true, Message: "系统已重置，减压阀和泵管减压阀已关闭，靶室真空泵已关闭"}
	}

}

// 发射控制
func (g *GasGun2Controller) Fire() APIResponse {
	err := g.OpenSwitch("FireSwitch")
	if err != nil {
		return APIResponse{Status: false, Message: "发射失败"}
	}

	go func() {
		time.Sleep(500 * time.Millisecond)
		g.CloseSwitch("FireSwitch")
	}()

	return APIResponse{Status: true, Message: "发射指令已执行"}
}

// 辅助函数：获取泵管压力（从PLC读取真实值）
func (g *GasGun2Controller) getPumpTubePressure() float32 {
	if g.xinjieClient == nil || g.xinjieClient.Client == nil {
		return 0
	}

	// 读取泵管压力（地址：VD4）
	values := g.xinjieClient.ReadFloat32(4, 1)
	if len(values) == 0 {
		fmt.Println("Failed to read pump tube pressure")
		return 0
	}

	return values[0]
}

// 辅助函数：获取气瓶压力（从PLC读取真实值）
func (g *GasGun2Controller) getCylinderPressure() float32 {
	if g.xinjieClient == nil || g.xinjieClient.Client == nil {
		return 0
	}

	// 读取气瓶压力（地址：VD2）
	values := g.xinjieClient.ReadFloat32(2, 1)
	if len(values) == 0 {
		fmt.Println("Failed to read cylinder pressure")
		return 0
	}

	return values[0]
}

// 手动操作函数
func (g *GasGun2Controller) ManualPumpTubePressurize(enable bool) APIResponse {
	if enable {
		err := g.OpenSwitch("PumpTubePressurize")
		if err != nil {
			return APIResponse{Status: false, Message: "打开泵管增压阀失败"}
		}
		return APIResponse{Status: true, Message: "泵管增压阀已打开"}
	} else {
		err := g.CloseSwitch("PumpTubePressurize")
		if err != nil {
			return APIResponse{Status: false, Message: "关闭泵管增压阀失败"}
		}
		return APIResponse{Status: true, Message: "泵管增压阀已关闭"}
	}
}

func (g *GasGun2Controller) ManualPumpTubeDecompress(enable bool) APIResponse {
	if enable {
		err := g.OpenSwitch("PumpTubeDecompress")
		if err != nil {
			return APIResponse{Status: false, Message: "打开泵管减压阀失败"}
		}
		return APIResponse{Status: true, Message: "泵管减压阀已打开"}
	} else {
		err := g.CloseSwitch("PumpTubeDecompress")
		if err != nil {
			return APIResponse{Status: false, Message: "关闭泵管减压阀失败"}
		}
		return APIResponse{Status: true, Message: "泵管减压阀已关闭"}
	}
}

func (g *GasGun2Controller) ManualPressurize(enable bool) APIResponse {
	if enable {
		err := g.OpenSwitch("Pressurize")
		if err != nil {
			return APIResponse{Status: false, Message: "打开增压阀失败"}
		}
		return APIResponse{Status: true, Message: "增压阀已打开"}
	} else {
		err := g.CloseSwitch("Pressurize")
		if err != nil {
			return APIResponse{Status: false, Message: "关闭增压阀失败"}
		}
		return APIResponse{Status: true, Message: "增压阀已关闭"}
	}
}

func (g *GasGun2Controller) ManualDecompress(enable bool) APIResponse {
	if enable {
		err := g.OpenSwitch("Decompress")
		if err != nil {
			return APIResponse{Status: false, Message: "打开减压阀失败"}
		}
		return APIResponse{Status: true, Message: "减压阀已打开"}
	} else {
		err := g.CloseSwitch("Decompress")
		if err != nil {
			return APIResponse{Status: false, Message: "关闭减压阀失败"}
		}
		return APIResponse{Status: true, Message: "减压阀已关闭"}
	}
}

func (g *GasGun2Controller) ManualPumpTubeProtect(enable bool) APIResponse {
	if enable {
		err := g.OpenSwitch("PumpTubeProtect")
		if err != nil {
			return APIResponse{Status: false, Message: "打开泵管保护阀失败"}
		}
		return APIResponse{Status: true, Message: "泵管保护阀已打开"}
	} else {
		err := g.CloseSwitch("PumpTubeProtect")
		if err != nil {
			return APIResponse{Status: false, Message: "关闭泵管保护阀失败"}
		}
		return APIResponse{Status: true, Message: "泵管保护阀已关闭"}
	}
}
