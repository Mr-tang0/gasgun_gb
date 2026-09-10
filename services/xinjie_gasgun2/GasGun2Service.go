package xinjiegasgun2

import (
	"context"
	"errors"
	"fmt"
	services "gasgun_gb/services/tools"
	"sync"
	"time"

	XinJie "github.com/Mr-tang0/PIMSGoMod/protocol/XinJie"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type XinjieGasGun2 struct {
	plc            *XinJie.XinjieClient
	ctx            context.Context
	heartbeat_info *GasGun2Heartbeat
	start_time     time.Time

	stopListen context.CancelFunc // 用于停止轮询
	config     Config             // 配置结构体

	AutoPumpTubePressureFlag bool
	AutoCylinderPressureFlag bool

	config_path string
	log_path    string

	mu sync.RWMutex // 保护 Output map 的并发访问
}

func NewXinjieGasGun2() *XinjieGasGun2 {
	return &XinjieGasGun2{
		plc:    XinJie.NewXinjieClient("XL5E"),
		config: NewConfig(),
		heartbeat_info: &GasGun2Heartbeat{
			Output: map[string]bool{
				"Pressurize":         false,
				"Decompress":         false,
				"PumpTubePressurize": false,
				"PumpTubeDecompress": false,
				"PumpTubeVacuum":     false,
				"TargetVacuum":       false,
				"TailVacuumProtect":  false,
				"PumpTubeProtect":    false,
				"FireSwitch":         false,
				"SystemDecompress":   false,
				"TargetVacuumPump":   false,
				"TailVacuumPump":     false,
			},
		},
	}
}

func (this *XinjieGasGun2) Startup(ctx context.Context) {
	this.ctx = ctx
	this.start_time = time.Now()

	this.log_path = services.GetLogPath("PIMS", "GASGUN", "xinjie_gasgun_2")
	this.config_path = services.GetConfigPath("PIMS", "GASGUN", "xinjie_gasgun_2")

	fmt.Println(this.config_path, this.log_path)

	this.config.LoadLocalConfig(this.config_path)
}

func (this *XinjieGasGun2) GetConfig() Config {
	return this.config
}

func (this *XinjieGasGun2) SaveConfig(config Config) error {
	this.config = config
	return this.config.SaveLocalConfig(this.config_path)
}

func (this *XinjieGasGun2) ConnectDevice(ip string) error {
	if ip == "" {
		return errors.New("ip is empty")
	}
	services.Log(fmt.Sprintf("连接设备: %s", ip), this.log_path)
	err := this.plc.OpenTCP(ip, 1)
	if err != nil {
		services.Log(fmt.Sprintf("连接设备失败: %v", err), this.log_path)
		return err
	}
	services.Log("设备连接成功", this.log_path)

	this.heartbeat_info.Running = true
	runtime.EventsEmit(this.ctx, "xinjie_gasgun2_heartbeat", this.heartbeat_info)

	go this.heartbeat()
	return nil
}

func (this *XinjieGasGun2) DisconnectDevice() {
	services.Log("断开设备连接", this.log_path)
	this.plc.Close()
	this.heartbeat_info.Running = false
	runtime.EventsEmit(this.ctx, "xinjie_gasgun2_heartbeat", this.heartbeat_info)
}

func (this *XinjieGasGun2) heartbeat() {

	for {
		if !this.heartbeat_info.Running {
			services.Log("心跳包发生错误", this.log_path)
			this.DisconnectDevice()
			return
		}
		this.plc.WriteCoil(XinJie.M(0), true)

		this.heartbeat_info.Time = time.Now().Format("15:04:05")

		if v, err := this.plc.ReadRegister(XinJie.D(this.config.Data.InputPressure), XinJie.Float32, false); err == nil {
			this.heartbeat_info.InputPressure = v.(float32)
		}
		if v, err := this.plc.ReadRegister(XinJie.D(this.config.Data.CylinderPressure), XinJie.Float32, false); err == nil {
			const alpha = 0.3
			if this.heartbeat_info.CylinderPressure == 0 {
				this.heartbeat_info.CylinderPressure = v.(float32)
			} else {
				this.heartbeat_info.CylinderPressure += alpha * (v.(float32) - this.heartbeat_info.CylinderPressure)
			}
		}
		if v, err := this.plc.ReadRegister(XinJie.D(this.config.Data.PumpTubePressure), XinJie.Float32, false); err == nil {
			const alpha = 0.3
			if this.heartbeat_info.PumpTubePressure == 0 {
				this.heartbeat_info.PumpTubePressure = v.(float32)
			} else {
				this.heartbeat_info.PumpTubePressure += alpha * (v.(float32) - this.heartbeat_info.PumpTubePressure)
			}
		}
		if v, err := this.plc.ReadRegister(XinJie.D(this.config.Data.PumpTubePressureHi), XinJie.Float32, false); err == nil {
			this.heartbeat_info.PumpTubePressureHi = v.(float32)
		}
		if v, err := this.plc.ReadRegister(XinJie.D(this.config.Data.TargetVacuumDegree), XinJie.Float32, false); err == nil {
			this.heartbeat_info.TargetVacuumDegree = v.(float32)
		}
		if v, err := this.plc.ReadRegister(XinJie.D(this.config.Data.TailVacuumDegree), XinJie.Float32, false); err == nil {
			this.heartbeat_info.TailVacuumDegree = v.(float32)
		}
		Y, err := this.plc.ReadCoils(XinJie.Y(0), 16, false)
		if err == nil && len(Y) >= 16 {
			// 信捷Y编号为八进制：Y0-Y7之后直接是Y10-Y17（Y8/Y9不存在）
			// ReadCoils返回的是连续数组，在索引8处补2个占位(false)，使数组下标与Y编号一致
			// 即 aligned[10]=Y10, aligned[13]=Y13 ...
			aligned := make([]bool, 18)
			copy(aligned[0:8], Y[0:8])
			copy(aligned[10:18], Y[8:16])

			s := this.config.Switch
			this.mu.Lock()
			this.heartbeat_info.Output["Pressurize"] = aligned[s.Pressurize]
			this.heartbeat_info.Output["Decompress"] = aligned[s.Decompress]
			this.heartbeat_info.Output["PumpTubePressurize"] = aligned[s.PumpTubePressurize]
			this.heartbeat_info.Output["PumpTubeDecompress"] = aligned[s.PumpTubeDecompress]
			this.heartbeat_info.Output["PumpTubeVacuum"] = aligned[s.PumpTubeVacuum]
			this.heartbeat_info.Output["TargetVacuum"] = aligned[s.TargetVacuum]
			this.heartbeat_info.Output["TailVacuumProtect"] = aligned[s.TailVacuumProtect]
			this.heartbeat_info.Output["PumpTubeProtect"] = aligned[s.PumpTubeProtect]
			this.heartbeat_info.Output["FireSwitch"] = aligned[s.FireSwitch]
			this.heartbeat_info.Output["SystemDecompress"] = aligned[s.SystemDecompress]
			this.heartbeat_info.Output["TailVacuumPump"] = aligned[s.TailVacuumPump]
			this.heartbeat_info.Output["TargetVacuumPump"] = aligned[s.TargetVacuumPump]
			this.mu.Unlock()
		}

		runtime.EventsEmit(this.ctx, "xinjie_gasgun2_heartbeat", this.heartbeat_info)
		time.Sleep(100 * time.Millisecond)
	}
}

// 设置触发模式：外触发开14和15引脚，内触发则关闭14和15引脚，同时均关闭发射引脚
func (this *XinjieGasGun2) SetTriggerMode(isExternal bool) error {
	mode := "内触发"
	if isExternal {
		mode = "外触发"
	}
	services.Log(fmt.Sprintf("设置触发模式: %s", mode), this.log_path)
	if isExternal {
		err := this.plc.WriteCoils(XinJie.Y(15), []bool{true, true})
		if err != nil {
			services.Log(fmt.Sprintf("设置触发模式失败: %v", err), this.log_path)
			return err
		}
	} else {
		err := this.plc.WriteCoils(XinJie.Y(15), []bool{false, false})
		if err != nil {
			services.Log(fmt.Sprintf("设置触发模式失败: %v", err), this.log_path)
			return err
		}
	}
	return nil
}

func (this *XinjieGasGun2) getSwitchAddress(addr string) uint16 {
	switch addr {
	case "Pressurize":
		return this.config.Switch.Pressurize
	case "Decompress":
		return this.config.Switch.Decompress
	case "FireSwitch":
		return this.config.Switch.FireSwitch
	case "PumpTubePressurize":
		return this.config.Switch.PumpTubePressurize
	case "PumpTubeDecompress":
		return this.config.Switch.PumpTubeDecompress
	case "PumpTubeVacuum":
		return this.config.Switch.PumpTubeVacuum
	case "TargetVacuum":
		return this.config.Switch.TargetVacuum
	case "TailVacuumProtect":
		return this.config.Switch.TailVacuumProtect
	case "PumpTubeProtect":
		return this.config.Switch.PumpTubeProtect
	case "SystemDecompress":
		return this.config.Switch.SystemDecompress
	case "TailVacuumPump":
		return this.config.Switch.TailVacuumPump
	case "TargetVacuumPump":
		return this.config.Switch.TargetVacuumPump
	default:
		return 99
	}
}

// 关闭阀门
func (this *XinjieGasGun2) CloseSwitch(addr string) error {
	services.Log(fmt.Sprintf("关闭阀门: %s, 地址: %d", addr, this.getSwitchAddress(addr)), this.log_path)
	return this.plc.WriteCoil(XinJie.Y(this.getSwitchAddress(addr)), false)
}

// 打开阀门
func (this *XinjieGasGun2) OpenSwitch(addr string) error {
	services.Log(fmt.Sprintf("打开阀门: %s, 地址: %d", addr, this.getSwitchAddress(addr)), this.log_path)
	return this.plc.WriteCoil(XinJie.Y(this.getSwitchAddress(addr)), true)
}

// 1. 抽真空函数：开尾部真空泵-间隔1s-开尾部真空保护阀-间隔1s-开靶室真空泵
func (this *XinjieGasGun2) StartAutoVacuum() error {
	services.Log("开始抽真空", this.log_path)
	// 开尾部真空泵
	err := this.OpenSwitch("TailVacuumPump")
	if err != nil {
		services.Log(fmt.Sprintf("打开尾部真空泵失败: %v", err), this.log_path)
		return fmt.Errorf("打开尾部真空泵失败: %w", err)
	}

	time.Sleep(1 * time.Second)

	// 开尾部真空保护阀
	err = this.OpenSwitch("TailVacuumProtect")
	if err != nil {
		services.Log(fmt.Sprintf("打开尾部真空保护阀失败: %v", err), this.log_path)
		this.CloseSwitch("TailVacuumPump")
		return fmt.Errorf("打开尾部真空保护阀失败: %w", err)
	}

	time.Sleep(1 * time.Second)

	// 开靶室真空泵
	err = this.OpenSwitch("TargetVacuumPump")
	if err != nil {
		services.Log(fmt.Sprintf("打开靶室真空泵失败: %v", err), this.log_path)
		this.CloseSwitch("TailVacuumPump")
		return fmt.Errorf("打开靶室真空泵失败: %w", err)
	}

	services.Log("抽真空启动完成", this.log_path)
	return nil
}

// 停止抽真空
func (this *XinjieGasGun2) StopAutoVacuum() error {
	services.Log("停止抽真空", this.log_path)
	err := this.CloseSwitch("TailVacuumPump")
	if err != nil {
		services.Log(fmt.Sprintf("关闭尾部真空泵失败: %v", err), this.log_path)
		return fmt.Errorf("关闭尾部真空泵失败: %w", err)
	}

	time.Sleep(1 * time.Second)

	err = this.CloseSwitch("TailVacuumProtect")
	if err != nil {
		services.Log(fmt.Sprintf("关闭尾部真空保护阀失败: %v", err), this.log_path)
		return fmt.Errorf("关闭尾部真空保护阀失败: %w", err)
	}

	time.Sleep(1 * time.Second)

	err = this.CloseSwitch("TargetVacuumPump")
	if err != nil {
		services.Log(fmt.Sprintf("关闭靶室真空泵失败: %v", err), this.log_path)
		return fmt.Errorf("关闭靶室真空泵失败: %w", err)
	}
	services.Log("抽真空停止完成", this.log_path)
	return nil
}

// 2. 抽泵管函数：打开抽泵管真空阀
func (this *XinjieGasGun2) StartPumpTubeVacuum() error {
	services.Log("开始抽泵管", this.log_path)
	err := this.OpenSwitch("PumpTubeVacuum")
	if err != nil {
		services.Log(fmt.Sprintf("打开抽泵管真空阀失败: %v", err), this.log_path)
		return fmt.Errorf("打开抽泵管真空阀失败: %w", err)
	}
	return nil
}

func (this *XinjieGasGun2) StopPumpTubeVacuum() error {
	services.Log("停止抽泵管", this.log_path)
	err := this.CloseSwitch("PumpTubeVacuum")
	if err != nil {
		services.Log(fmt.Sprintf("关闭抽泵管真空阀失败: %v", err), this.log_path)
		return fmt.Errorf("关闭抽泵管真空阀失败: %w", err)
	}
	return nil
}

// 3. 子线程自动设置泵管压力（通过开关泵管增压阀与泵管减压阀）
func (g *XinjieGasGun2) AutoPumpTubePressure(target float32) {
	services.Log(fmt.Sprintf("自动设置泵管压力: %f", target), g.log_path)
	g.AutoPumpTubePressureFlag = true

	g.OpenSwitch("PumpTubeProtect")

	go func(target float32) {
		for {
			if !g.AutoPumpTubePressureFlag {
				break
			}

			g.mu.RLock()
			pressurize := g.heartbeat_info.Output["PumpTubePressurize"]
			decompress := g.heartbeat_info.Output["PumpTubeDecompress"]
			g.mu.RUnlock()

			if g.heartbeat_info.PumpTubePressureHi < target-0.02 {
				// 需要增压：打开增压阀，关闭减压阀
				if !pressurize {
					g.OpenSwitch("PumpTubePressurize")
				}
				if decompress {
					g.CloseSwitch("PumpTubeDecompress")
				}
			} else if g.heartbeat_info.PumpTubePressureHi > target+0.02 {
				// 需要减压：打开减压阀，关闭增压阀
				if !decompress {
					g.OpenSwitch("PumpTubeDecompress")
				}
				if pressurize {
					g.CloseSwitch("PumpTubePressurize")
				}
			} else {
				// 达到目标压力：关闭两个阀门
				if pressurize {
					g.CloseSwitch("PumpTubePressurize")
				}
				if decompress {
					g.CloseSwitch("PumpTubeDecompress")
				}
			}

			time.Sleep(100 * time.Millisecond)

		}
	}(target)

}

func (g *XinjieGasGun2) StopAutoPumpTubePressure() {
	services.Log("停止自动泵管压力控制", g.log_path)
	g.AutoPumpTubePressureFlag = false
	// 关闭泵管增压阀与泵管减压阀
	g.CloseSwitch("PumpTubePressurize")
	g.CloseSwitch("PumpTubeDecompress")
	g.CloseSwitch("PumpTubeProtect")
}

// 4. 子线程自动设置气瓶压力（通过开关增压阀与减压阀）
func (g *XinjieGasGun2) AutoCylinderPressure(target float32) {
	services.Log(fmt.Sprintf("自动设置气瓶压力: %f", target), g.log_path)
	g.AutoCylinderPressureFlag = true

	go func(target float32) {
		for {
			if !g.AutoCylinderPressureFlag {
				break
			}

			g.mu.RLock()
			pressurize := g.heartbeat_info.Output["Pressurize"]
			decompress := g.heartbeat_info.Output["Decompress"]
			g.mu.RUnlock()

			if g.heartbeat_info.CylinderPressure < target-0.1 {
				// 需要增压：打开增压阀，关闭减压阀
				if !pressurize {
					g.OpenSwitch("Pressurize")
				}
				if decompress {
					g.CloseSwitch("Decompress")
				}
			} else if g.heartbeat_info.CylinderPressure > target+0.1 {
				// 需要减压：打开减压阀，关闭增压阀
				if !decompress {
					g.OpenSwitch("Decompress")
				}
				if pressurize {
					g.CloseSwitch("Pressurize")
				}
			} else {
				// 达到目标压力：关闭两个阀门
				if pressurize {
					g.CloseSwitch("Pressurize")
				}
				if decompress {
					g.CloseSwitch("Decompress")
				}
			}

			time.Sleep(100 * time.Millisecond)
		}
	}(target)

}

func (g *XinjieGasGun2) StopAutoCylinderPressure() {
	services.Log("停止自动气瓶压力控制", g.log_path)
	g.AutoCylinderPressureFlag = false
	// 关闭增瓶增压阀与增瓶减压阀
	g.CloseSwitch("Pressurize")
	g.CloseSwitch("Decompress")
}

// 5. 准备发射：关闭尾部真空泵，尾部真空阀，靶室真空泵，打开发射阀（内触发时）
func (this *XinjieGasGun2) PrepareFire() error {
	services.Log("准备发射", this.log_path)
	// 关闭尾部真空泵
	err := this.CloseSwitch("TailVacuumPump")
	if err != nil {
		services.Log(fmt.Sprintf("关闭尾部真空泵失败: %v", err), this.log_path)
		return fmt.Errorf("关闭尾部真空泵失败: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	// 关闭尾部真空阀
	err = this.CloseSwitch("TailVacuumProtect")
	if err != nil {
		services.Log(fmt.Sprintf("关闭尾部真空阀失败: %v", err), this.log_path)
		return fmt.Errorf("关闭尾部真空阀失败: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	// 关闭靶室真空泵
	err = this.CloseSwitch("TargetVacuumPump")
	if err != nil {
		services.Log(fmt.Sprintf("关闭靶室真空泵失败: %v", err), this.log_path)
		return fmt.Errorf("关闭靶室真空泵失败: %w", err)
	}
	services.Log("准备发射完成", this.log_path)
	return nil

}

// 6. 恢复：true:打开减压阀、泵管减压阀、系统减压阀 false:关闭减压阀、泵管减压阀、系统减压阀
func (this *XinjieGasGun2) ResetSystem(reset bool) error {
	if reset {
		services.Log("恢复系统: 打开减压阀、泵管减压阀、系统减压阀", this.log_path)
		// 打开减压阀
		err := this.OpenSwitch("Decompress")
		if err != nil {
			services.Log(fmt.Sprintf("打开减压阀失败: %v", err), this.log_path)
			return fmt.Errorf("打开减压阀失败: %w", err)
		}

		time.Sleep(1000 * time.Millisecond)

		// 打开泵管减压阀
		err = this.OpenSwitch("PumpTubeDecompress")
		if err != nil {
			services.Log(fmt.Sprintf("打开泵管减压阀失败: %v", err), this.log_path)
			return fmt.Errorf("打开泵管减压阀失败: %w", err)
		}

		time.Sleep(1000 * time.Millisecond)

		// 重新开启靶室真空泵
		err = this.OpenSwitch("SystemDecompress")
		if err != nil {
			services.Log(fmt.Sprintf("开启靶室真空泵失败: %v", err), this.log_path)
			return fmt.Errorf("开启靶室真空泵失败: %w", err)
		}

		services.Log("系统恢复完成", this.log_path)
		return nil
	} else {
		services.Log("恢复系统: 关闭减压阀、泵管减压阀、系统减压阀", this.log_path)
		// 关闭减压阀
		err := this.CloseSwitch("Decompress")
		if err != nil {
			services.Log(fmt.Sprintf("关闭减压阀失败: %v", err), this.log_path)
			return fmt.Errorf("关闭减压阀失败: %w", err)
		}

		time.Sleep(1000 * time.Millisecond)

		// 关闭泵管减压阀
		err = this.CloseSwitch("PumpTubeDecompress")
		if err != nil {
			services.Log(fmt.Sprintf("关闭泵管减压阀失败: %v", err), this.log_path)
			return fmt.Errorf("关闭泵管减压阀失败: %w", err)
		}

		time.Sleep(1000 * time.Millisecond)

		err = this.CloseSwitch("SystemDecompress")
		if err != nil {
			services.Log(fmt.Sprintf("关闭目标真空泵阀失败: %v", err), this.log_path)
			return fmt.Errorf("关闭目标真空泵阀失败: %w", err)
		}

		services.Log("系统关闭完成", this.log_path)
		return nil
	}

}

// 发射控制
func (this *XinjieGasGun2) Fire() error {
	services.Log("发射", this.log_path)
	err := this.OpenSwitch("FireSwitch")
	if err != nil {
		services.Log(fmt.Sprintf("发射失败: %v", err), this.log_path)
		return fmt.Errorf("发射失败")
	}

	go func() {
		time.Sleep(500 * time.Millisecond)
		this.CloseSwitch("FireSwitch")
	}()

	return nil
}
