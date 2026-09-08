package xinjiegasgun2

import (
	"context"
	"errors"
	"fmt"
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
}

func NewXinjieGasGun2() *XinjieGasGun2 {
	return &XinjieGasGun2{
		plc:    XinJie.NewXinjieClient("XL5E"),
		config: NewConfig(),
	}
}

func (this *XinjieGasGun2) Startup(ctx context.Context) {
	this.ctx = ctx
	this.start_time = time.Now()
	this.config.LoadLocalConfig()
}

func (this *XinjieGasGun2) ConnectDevice(ip string) error {
	if ip == "" {
		return errors.New("ip is empty")
	}
	err := this.plc.OpenTCP(ip, 1)
	if err != nil {
		return err
	}

	this.heartbeat_info.Running = true
	runtime.EventsEmit(this.ctx, "heartbeat", this.heartbeat_info)

	go this.heartbeat()
	return nil
}

func (this *XinjieGasGun2) DisconnectDevice() {
	this.plc.Close()
	this.heartbeat_info.Running = false
	runtime.EventsEmit(this.ctx, "heartbeat", this.heartbeat_info)
}

func (this *XinjieGasGun2) heartbeat() {
	for {
		if !this.heartbeat_info.Running {
			return
		}
		this.plc.WriteCoil(XinJie.M(0), true)

		time.Sleep(100 * time.Millisecond)
	}
}

// 设置触发模式：外触发开14和15引脚，内触发则关闭14和15引脚，同时均关闭发射引脚
func (this *XinjieGasGun2) SetTriggerMode(isExternal bool) error {
	err := this.plc.WriteCoil(XinJie.Y(this.config.Switch.FireSwitch), false)
	if err != nil {
		fmt.Printf("关闭发射阀失败: %v\n", err)
		return err
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

	default:
		return 0
	}
}

// 关闭阀门
func (this *XinjieGasGun2) CloseSwitch(addr string) error {
	return this.plc.WriteCoil(XinJie.Y(this.getSwitchAddress(addr)), false)
}

// 打开阀门
func (this *XinjieGasGun2) OpenSwitch(addr string) error {
	return this.plc.WriteCoil(XinJie.Y(this.getSwitchAddress(addr)), true)
}

// 1. 抽真空函数：开尾部真空泵-间隔1s-开尾部真空保护阀-间隔1s-开靶室真空泵
func (this *XinjieGasGun2) StartAutoVacuum() error {
	// 开尾部真空泵
	err := this.OpenSwitch("TailVacuumPump")
	if err != nil {
		return fmt.Errorf("打开尾部真空泵失败: %w", err)
	}

	time.Sleep(1 * time.Second)

	// 开尾部真空保护阀
	err = this.OpenSwitch("TailVacuumProtect")
	if err != nil {
		this.CloseSwitch("TailVacuumPump")
		return fmt.Errorf("打开尾部真空保护阀失败: %w", err)
	}

	time.Sleep(1 * time.Second)

	// 开靶室真空泵
	err = this.OpenSwitch("TargetVacuumPump")
	if err != nil {
		this.CloseSwitch("TailVacuumPump")
		return fmt.Errorf("打开靶室真空泵失败: %w", err)
	}

	return nil
}

// 停止抽真空
func (this *XinjieGasGun2) StopAutoVacuum() error {
	err := this.CloseSwitch("TailVacuumPump")
	if err != nil {
		return fmt.Errorf("关闭尾部真空泵失败: %w", err)
	}

	time.Sleep(1 * time.Second)

	err = this.CloseSwitch("TailVacuumProtect")
	if err != nil {
		return fmt.Errorf("关闭尾部真空保护阀失败: %w", err)
	}

	time.Sleep(1 * time.Second)

	err = this.CloseSwitch("TargetVacuumPump")
	if err != nil {
		return fmt.Errorf("关闭靶室真空泵失败: %w", err)
	}
	return nil
}

// 2. 抽泵管函数：打开抽泵管真空阀
func (this *XinjieGasGun2) StartPumpTubeVacuum() error {
	err := this.OpenSwitch("PumpTubeVacuum")
	if err != nil {
		return fmt.Errorf("打开抽泵管真空阀失败: %w", err)
	}
	return nil
}

func (this *XinjieGasGun2) StopPumpTubeVacuum() error {
	err := this.CloseSwitch("PumpTubeVacuum")
	if err != nil {
		return fmt.Errorf("关闭抽泵管真空阀失败: %w", err)
	}
	return nil
}

// 3. 子线程自动设置泵管压力（通过开关泵管增压阀与泵管减压阀）
func (g *XinjieGasGun2) AutoPumpTubePressure(target float32) {
	g.AutoPumpTubePressureFlag = true

	go func(target float32) {
		for {
			if !g.AutoPumpTubePressureFlag {
				break
			}

			if g.heartbeat_info.PumpTubePressure > target+0.1 {
				// 需要增压：打开增压阀，关闭减压阀
				g.OpenSwitch("PumpTubePressurize")
				g.CloseSwitch("PumpTubeDecompress")
			} else if g.heartbeat_info.PumpTubePressure < target-0.1 {
				// 需要减压：打开减压阀，关闭增压阀
				g.OpenSwitch("PumpTubeDecompress")
				g.CloseSwitch("PumpTubePressurize")
			} else {
				// 达到目标压力：关闭两个阀门
				g.CloseSwitch("PumpTubePressurize")
				g.CloseSwitch("PumpTubeDecompress")
			}

			time.Sleep(100 * time.Millisecond)

		}
	}(target)

}

func (g *XinjieGasGun2) StopAutoPumpTubePressure() {
	g.AutoPumpTubePressureFlag = false
}

// 4. 子线程自动设置气瓶压力（通过开关增压阀与减压阀）
func (g *XinjieGasGun2) AutoCylinderPressure(target float32) {
	g.AutoCylinderPressureFlag = true

	go func(target float32) {
		for {
			if !g.AutoCylinderPressureFlag {
				break
			}

			if g.heartbeat_info.CylinderPressure > target+0.1 {
				// 需要增压：打开增压阀，关闭减压阀
				g.OpenSwitch("CylinderPressurize")
				g.CloseSwitch("CylinderDecompress")
			} else if g.heartbeat_info.CylinderPressure < target-0.1 {
				// 需要减压：打开减压阀，关闭增压阀
				g.OpenSwitch("CylinderDecompress")
				g.CloseSwitch("CylinderPressurize")
			} else {
				// 达到目标压力：关闭两个阀门
				g.CloseSwitch("CylinderPressurize")
				g.CloseSwitch("CylinderDecompress")
			}

		}
	}(target)

}

func (g *XinjieGasGun2) StopAutoCylinderPressure() {
	g.AutoCylinderPressureFlag = false
}

// 5. 准备发射：关闭尾部真空泵，尾部真空阀，靶室真空泵，打开发射阀（内触发时）
func (this *XinjieGasGun2) PrepareFire() error {
	// 关闭尾部真空泵
	err := this.CloseSwitch("TailVacuumPump")
	if err != nil {
		return fmt.Errorf("关闭尾部真空泵失败: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	// 关闭尾部真空阀
	err = this.CloseSwitch("TailVacuumProtect")
	if err != nil {
		return fmt.Errorf("关闭尾部真空阀失败: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	// 关闭靶室真空泵
	err = this.CloseSwitch("TargetVacuumPump")
	if err != nil {
		return fmt.Errorf("关闭靶室真空泵失败: %w", err)
	}
	return nil

}

// 6. 恢复：true:打开减压阀、泵管减压阀、系统减压阀 false:关闭减压阀、泵管减压阀、系统减压阀
func (this *XinjieGasGun2) ResetSystem(reset bool) error {
	if reset {
		// 打开减压阀
		err := this.OpenSwitch("Decompress")
		if err != nil {
			return fmt.Errorf("打开减压阀失败: %w", err)
		}

		time.Sleep(500 * time.Millisecond)

		// 打开泵管减压阀
		err = this.OpenSwitch("PumpTubeDecompress")
		if err != nil {
			return fmt.Errorf("打开泵管减压阀失败: %w", err)
		}

		time.Sleep(500 * time.Millisecond)

		// 重新开启靶室真空泵
		err = this.OpenSwitch("SystemDecompress")
		if err != nil {
			return fmt.Errorf("开启靶室真空泵失败: %w", err)
		}

		return nil
	} else {
		// 关闭减压阀
		err := this.CloseSwitch("Decompress")
		if err != nil {
			return fmt.Errorf("关闭减压阀失败: %w", err)
		}

		time.Sleep(500 * time.Millisecond)

		// 关闭泵管减压阀
		err = this.CloseSwitch("PumpTubeDecompress")
		if err != nil {
			return fmt.Errorf("关闭泵管减压阀失败: %w", err)
		}

		time.Sleep(500 * time.Millisecond)

		err = this.CloseSwitch("SystemDecompress")
		if err != nil {
			return fmt.Errorf("关闭目标真空泵阀失败: %w", err)
		}

		return nil
	}

}

// 发射控制
func (this *XinjieGasGun2) Fire() error {
	err := this.OpenSwitch("FireSwitch")
	if err != nil {
		return fmt.Errorf("发射失败")
	}

	go func() {
		time.Sleep(500 * time.Millisecond)
		this.CloseSwitch("FireSwitch")
	}()

	return nil
}

// 手动操作函数
func (this *XinjieGasGun2) ManualPumpTubePressurize(enable bool) error {
	err := errors.New("")
	if enable {
		err = this.OpenSwitch("PumpTubePressurize")
	} else {
		err = this.CloseSwitch("PumpTubePressurize")
	}

	if err != nil {
		return fmt.Errorf("操作泵管增压阀失败: %w", err)
	}
	return nil
}

func (g *XinjieGasGun2) ManualPumpTubeDecompress(enable bool) error {
	err := errors.New("")
	if enable {
		err = g.OpenSwitch("PumpTubeDecompress")
	} else {
		err = g.CloseSwitch("PumpTubeDecompress")
	}

	if err != nil {
		return fmt.Errorf("操作泵管减压阀失败: %w", err)
	}
	return nil
}

func (g *XinjieGasGun2) ManualPressurize(enable bool) error {
	err := errors.New("")
	if enable {
		err = g.OpenSwitch("Pressurize")
	} else {
		err = g.CloseSwitch("Pressurize")
	}
	if err != nil {
		return fmt.Errorf("操作增压阀失败: %w", err)
	}
	return nil
}

func (g *XinjieGasGun2) ManualDecompress(enable bool) error {
	err := errors.New("")
	if enable {
		err = g.OpenSwitch("Decompress")
	} else {
		err = g.CloseSwitch("Decompress")

	}
	if err != nil {
		return fmt.Errorf("操作减压阀失败: %w", err)
	}
	return nil
}

func (g *XinjieGasGun2) ManualPumpTubeProtect(enable bool) error {
	err := errors.New("")
	if enable {
		err = g.OpenSwitch("PumpTubeProtect")
	} else {
		err = g.CloseSwitch("PumpTubeProtect")
	}
	if err != nil {
		return fmt.Errorf("操作泵管保护阀失败: %w", err)
	}
	return nil
}
