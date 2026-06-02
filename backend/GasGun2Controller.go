package backend

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type SWITCH struct {
	Pressurize         int `json:"Pressurize"`          //增压阀1
	Decompress         int `json:"Decompress"`          //减压阀1
	PumpTubePressurize int `json:"PumpTTubePressurize"` //泵管增压阀1
	PumpTubeDecompress int `json:"PumpTubeDecompress"`  //泵管减压阀1
	PumpTubeVacuum     int `json:"PumpTubeVacuum"`      //抽泵管真空阀1
	// PumpTubeVacuumRelease int `json:"PumpTubeRelease"`     //泵管真空释放阀
	TargetVacuum int `json:"TargetVacuum"` //靶室真空阀1
	// TargetVacuumRelease int `json:"TargetVacuumRelease"` //靶室真空释放阀
	TailVacuumProtect int `json:"TailVacuumProtect"` //尾真空保护阀（尾部真空阀）
	PumpTubeProtect   int `json:"PumpTubeProtect"`   //泵管保护阀1
	FireSwitch        int `json:"FireSwitch"`        //发射阀1
	SystemDecompress  int `json:"SystemDecompress"`  //系统减压阀1
	TargetVacuumPump  int `json:"TargetVacuumpump"`  //靶室真空泵1
	TailVacuumPump    int `json:"TailVacuumPump"`    //尾真空泵1
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
	PLCSwitch  SWITCH             // 存储配置

	// 自动压力控制相关
	pumpTubeTargetPressure float32
	cylinderTargetPressure float32
	stopPumpTubeControl    context.CancelFunc
	stopCylinderControl    context.CancelFunc
	isExternalTrigger      bool // 是否为外触发模式
}

func NewGasGun2Controller() *GasGun2Controller {
	return &GasGun2Controller{}
}

func (g *GasGun2Controller) Init(ctx context.Context) {
	g.ctx = ctx
}

func (g *GasGun2Controller) GetConfig() SWITCH {
	return g.PLCSwitch
}

func (g *GasGun2Controller) SaveConfig(newConfig SWITCH) {
	g.PLCSwitch = newConfig
}

// 设置触发模式
func (g *GasGun2Controller) SetTriggerMode(isExternal bool) {
	g.isExternalTrigger = isExternal
}

// *********************************** PLC通信区域 *********************************

func (g *GasGun2Controller) ConnectPLC(ip string) APIResponse {
	fmt.Println("GasGun2 connecting to:", ip)

	// 模拟连接成功
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

	return APIResponse{Status: true, Message: "已断开连接并停止采集！"}
}

// 子线程获取PLC监控参数并实时向前端发送数据
func (g *GasGun2Controller) GetRealTimeData() {
	// 模拟从PLC读取数据（实际项目中需要调用ReadFloat32等方法）
	metrics := GasGun2Metrics{
		InputPressure:      0.85,   // 输入压力
		CylinderPressure:   12.3,   // 气瓶压力（一级气室）
		PumpTubePressure:   25.6,   // 泵管压力（二级气室）
		PumpTubePressureHi: 25.634, // 泵管压力（高精度）
		TargetVacuumDegree: 100.5,  // 靶室真空度
		TailVacuumDegree:   150.2,  // 尾部真空度
	}

	// 发送数据到前端
	runtime.EventsEmit(g.ctx, "update_gasgun2_metrics", metrics)
}

// *********************************** 控制阀门区域 *********************************

func (g *GasGun2Controller) OpenSwitch(s int) error {
	fmt.Printf("打开阀门: %d\n", s)
	return nil
}

func (g *GasGun2Controller) CloseSwitch(s int) error {
	fmt.Printf("关闭阀门: %d\n", s)
	return nil
}

//****************************************自动化流程*****************************************

// 1. 抽真空函数：开尾部真空泵-间隔1s-开尾部真空保护阀-间隔1s-开靶室真空泵
func (g *GasGun2Controller) StartAutoVacuum() APIResponse {
	// 开尾部真空泵
	err := g.OpenSwitch(g.PLCSwitch.TailVacuumPump)
	if err != nil {
		return APIResponse{Status: false, Message: "打开尾部真空泵失败"}
	}

	time.Sleep(1 * time.Second)

	// 开尾部真空保护阀
	err = g.OpenSwitch(g.PLCSwitch.TailVacuumProtect)
	if err != nil {
		g.CloseSwitch(g.PLCSwitch.TailVacuumPump)
		return APIResponse{Status: false, Message: "打开尾部真空保护阀失败"}
	}

	time.Sleep(1 * time.Second)

	// 开靶室真空泵
	err = g.OpenSwitch(g.PLCSwitch.TargetVacuumPump)
	if err != nil {
		g.CloseSwitch(g.PLCSwitch.TailVacuumPump)
		g.CloseSwitch(g.PLCSwitch.TailVacuumProtect)
		return APIResponse{Status: false, Message: "打开靶室真空泵失败"}
	}

	return APIResponse{Status: true, Message: "抽真空流程已启动"}
}

// 停止抽真空
func (g *GasGun2Controller) StopAutoVacuum() APIResponse {
	err := g.CloseSwitch(g.PLCSwitch.TailVacuumPump)
	if err != nil {
		return APIResponse{Status: false, Message: "关闭尾部真空泵失败"}
	}

	time.Sleep(1 * time.Second)

	err = g.CloseSwitch(g.PLCSwitch.TailVacuumProtect)
	if err != nil {
		return APIResponse{Status: false, Message: "关闭尾部真空保护阀失败"}
	}

	time.Sleep(1 * time.Second)

	err = g.CloseSwitch(g.PLCSwitch.TargetVacuumPump)
	if err != nil {
		return APIResponse{Status: false, Message: "关闭靶室真空泵失败"}
	}

	return APIResponse{Status: true, Message: "抽真空流程已停止"}
}

// 2. 抽泵管函数：打开抽泵管真空阀
func (g *GasGun2Controller) StartPumpTubeVacuum() APIResponse {
	err := g.OpenSwitch(g.PLCSwitch.PumpTubeVacuum)
	if err != nil {
		return APIResponse{Status: false, Message: "打开抽泵管真空阀失败"}
	}
	return APIResponse{Status: true, Message: "抽泵管真空已启动"}
}

func (g *GasGun2Controller) StopPumpTubeVacuum() APIResponse {
	err := g.CloseSwitch(g.PLCSwitch.PumpTubeVacuum)
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
				// 获取当前泵管压力（模拟）
				currentPressure := g.getPumpTubePressure()

				// 控制逻辑：根据当前压力与目标压力的差值来控制阀门
				if math.Abs(float64(currentPressure-target)) > 0.1 {
					if currentPressure < target {
						// 需要增压：打开增压阀，关闭减压阀
						g.OpenSwitch(g.PLCSwitch.PumpTubePressurize)
						g.CloseSwitch(g.PLCSwitch.PumpTubeDecompress)
					} else {
						// 需要减压：打开减压阀，关闭增压阀
						g.OpenSwitch(g.PLCSwitch.PumpTubeDecompress)
						g.CloseSwitch(g.PLCSwitch.PumpTubePressurize)
					}
				} else {
					// 达到目标压力：关闭两个阀门
					g.CloseSwitch(g.PLCSwitch.PumpTubePressurize)
					g.CloseSwitch(g.PLCSwitch.PumpTubeDecompress)
					fmt.Printf("泵管压力已达到目标值: %.2f\n", target)
					return
				}
			case <-ctx.Done():
				// 停止时关闭阀门
				g.CloseSwitch(g.PLCSwitch.PumpTubePressurize)
				g.CloseSwitch(g.PLCSwitch.PumpTubeDecompress)
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
				// 获取当前气瓶压力（模拟）
				currentPressure := g.getCylinderPressure()

				// 控制逻辑
				if math.Abs(float64(currentPressure-target)) > 0.1 {
					if currentPressure < target {
						// 需要增压
						g.OpenSwitch(g.PLCSwitch.Pressurize)
						g.CloseSwitch(g.PLCSwitch.Decompress)
					} else {
						// 需要减压
						g.OpenSwitch(g.PLCSwitch.Decompress)
						g.CloseSwitch(g.PLCSwitch.Pressurize)
					}
				} else {
					// 达到目标压力
					g.CloseSwitch(g.PLCSwitch.Pressurize)
					g.CloseSwitch(g.PLCSwitch.Decompress)
					fmt.Printf("气瓶压力已达到目标值: %.2f\n", target)
					return
				}
			case <-ctx.Done():
				g.CloseSwitch(g.PLCSwitch.Pressurize)
				g.CloseSwitch(g.PLCSwitch.Decompress)
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
	err := g.CloseSwitch(g.PLCSwitch.TailVacuumPump)
	if err != nil {
		return APIResponse{Status: false, Message: "关闭尾部真空泵失败"}
	}

	time.Sleep(500 * time.Millisecond)

	// 关闭尾部真空阀
	err = g.CloseSwitch(g.PLCSwitch.TailVacuumProtect)
	if err != nil {
		return APIResponse{Status: false, Message: "关闭尾部真空阀失败"}
	}

	time.Sleep(500 * time.Millisecond)

	// 关闭靶室真空泵
	err = g.CloseSwitch(g.PLCSwitch.TargetVacuumPump)
	if err != nil {
		return APIResponse{Status: false, Message: "关闭靶室真空泵失败"}
	}

	time.Sleep(500 * time.Millisecond)

	// 判断是内触发还是外触发，外触发时不打开发射阀
	if !g.isExternalTrigger {
		// 内触发：打开发射阀
		err = g.OpenSwitch(g.PLCSwitch.FireSwitch)
		if err != nil {
			return APIResponse{Status: false, Message: "打开发射阀失败"}
		}
		return APIResponse{Status: true, Message: "准备发射完成（内触发模式，发射阀已打开）"}
	} else {
		return APIResponse{Status: true, Message: "准备发射完成（外触发模式，等待外部触发信号）"}
	}
}

// 6. 恢复：打开减压阀、泵管减压阀、重新开启靶室真空泵
func (g *GasGun2Controller) ResetSystem() APIResponse {
	// 打开减压阀
	err := g.OpenSwitch(g.PLCSwitch.Decompress)
	if err != nil {
		return APIResponse{Status: false, Message: "打开减压阀失败"}
	}

	time.Sleep(500 * time.Millisecond)

	// 打开泵管减压阀
	err = g.OpenSwitch(g.PLCSwitch.PumpTubeDecompress)
	if err != nil {
		return APIResponse{Status: false, Message: "打开泵管减压阀失败"}
	}

	time.Sleep(500 * time.Millisecond)

	// 重新开启靶室真空泵
	err = g.OpenSwitch(g.PLCSwitch.TargetVacuumPump)
	if err != nil {
		return APIResponse{Status: false, Message: "开启靶室真空泵失败"}
	}

	return APIResponse{Status: true, Message: "系统恢复完成"}
}

// 发射控制
func (g *GasGun2Controller) Fire() APIResponse {
	err := g.OpenSwitch(g.PLCSwitch.FireSwitch)
	if err != nil {
		return APIResponse{Status: false, Message: "发射失败"}
	}

	go func() {
		time.Sleep(500 * time.Millisecond)
		g.CloseSwitch(g.PLCSwitch.FireSwitch)
	}()

	return APIResponse{Status: true, Message: "发射指令已执行"}
}

// 辅助函数：获取泵管压力（模拟）
func (g *GasGun2Controller) getPumpTubePressure() float32 {
	return g.pumpTubeTargetPressure * 0.8 // 模拟值
}

// 辅助函数：获取气瓶压力（模拟）
func (g *GasGun2Controller) getCylinderPressure() float32 {
	return g.cylinderTargetPressure * 0.7 // 模拟值
}

// 手动操作函数
func (g *GasGun2Controller) ManualPumpTubePressurize(enable bool) APIResponse {
	if enable {
		err := g.OpenSwitch(g.PLCSwitch.PumpTubePressurize)
		if err != nil {
			return APIResponse{Status: false, Message: "打开泵管增压阀失败"}
		}
		return APIResponse{Status: true, Message: "泵管增压阀已打开"}
	} else {
		err := g.CloseSwitch(g.PLCSwitch.PumpTubePressurize)
		if err != nil {
			return APIResponse{Status: false, Message: "关闭泵管增压阀失败"}
		}
		return APIResponse{Status: true, Message: "泵管增压阀已关闭"}
	}
}

func (g *GasGun2Controller) ManualPumpTubeDecompress(enable bool) APIResponse {
	if enable {
		err := g.OpenSwitch(g.PLCSwitch.PumpTubeDecompress)
		if err != nil {
			return APIResponse{Status: false, Message: "打开泵管减压阀失败"}
		}
		return APIResponse{Status: true, Message: "泵管减压阀已打开"}
	} else {
		err := g.CloseSwitch(g.PLCSwitch.PumpTubeDecompress)
		if err != nil {
			return APIResponse{Status: false, Message: "关闭泵管减压阀失败"}
		}
		return APIResponse{Status: true, Message: "泵管减压阀已关闭"}
	}
}

func (g *GasGun2Controller) ManualPressurize(enable bool) APIResponse {
	if enable {
		err := g.OpenSwitch(g.PLCSwitch.Pressurize)
		if err != nil {
			return APIResponse{Status: false, Message: "打开增压阀失败"}
		}
		return APIResponse{Status: true, Message: "增压阀已打开"}
	} else {
		err := g.CloseSwitch(g.PLCSwitch.Pressurize)
		if err != nil {
			return APIResponse{Status: false, Message: "关闭增压阀失败"}
		}
		return APIResponse{Status: true, Message: "增压阀已关闭"}
	}
}

func (g *GasGun2Controller) ManualDecompress(enable bool) APIResponse {
	if enable {
		err := g.OpenSwitch(g.PLCSwitch.Decompress)
		if err != nil {
			return APIResponse{Status: false, Message: "打开减压阀失败"}
		}
		return APIResponse{Status: true, Message: "减压阀已打开"}
	} else {
		err := g.CloseSwitch(g.PLCSwitch.Decompress)
		if err != nil {
			return APIResponse{Status: false, Message: "关闭减压阀失败"}
		}
		return APIResponse{Status: true, Message: "减压阀已关闭"}
	}
}

func (g *GasGun2Controller) ManualPumpTubeProtect(enable bool) APIResponse {
	if enable {
		err := g.OpenSwitch(g.PLCSwitch.PumpTubeProtect)
		if err != nil {
			return APIResponse{Status: false, Message: "打开泵管保护阀失败"}
		}
		return APIResponse{Status: true, Message: "泵管保护阀已打开"}
	} else {
		err := g.CloseSwitch(g.PLCSwitch.PumpTubeProtect)
		if err != nil {
			return APIResponse{Status: false, Message: "关闭泵管保护阀失败"}
		}
		return APIResponse{Status: true, Message: "泵管保护阀已关闭"}
	}
}
