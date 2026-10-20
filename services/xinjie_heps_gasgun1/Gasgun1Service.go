package hepsgasgun1

import (
	"context"
	"errors"
	"fmt"
	services "gasgun_gb/services/tools"
	"os"
	"sync"
	"time"

	XinJie "github.com/Mr-tang0/PIMSGoMod/protocol/XinJie"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type HEPSGasGun1 struct {
	plc            *XinJie.XinjieClient
	ctx            context.Context
	heartbeat_info *HEPSGasgun1Heartbeat

	start_time time.Time

	config      Config
	config_path string
	log_path    string

	AutoPressureFlag bool // 自动压力控制运行标志

	mu sync.RWMutex // 保护心跳状态的并发访问
}

func NewHEPSGasGun1() *HEPSGasGun1 {
	return &HEPSGasGun1{
		plc:    XinJie.NewXinjieClient("XD"),
		config: NewConfig(),
		heartbeat_info: &HEPSGasgun1Heartbeat{
			Running: false,
			Alarm:   "",
			Time:    "",

			InputPressure:      0,
			CylinderPressure:   0,
			TargetVacuumDegree: 0,
			TailVacuumDegree:   0,

			Output: map[string]bool{
				"Fire":   false,
				"Inlet":  false,
				"Outlet": false,

				"TailVacuumPump":      false,
				"TargetVacuumRelease": false,
				"TarVacuumPump":       false,
				"SystemDecompress":    false,
			},
			Input: map[string]bool{
				"FireTrigger": false,
			},
		},
	}
}

func (this *HEPSGasGun1) Startup(ctx context.Context) {
	this.ctx = ctx
	this.start_time = time.Now()
	this.log_path = services.GetLogPath("PIMS", "GASGUN", "xinjie_heps_gasgun_1")
	this.config_path = services.GetConfigPath("PIMS", "GASGUN", "xinjie_heps_gasgun_1")

	fmt.Println(this.config_path, this.log_path)
	this.config.LoadLocalConfig(this.config_path)
}

func (this *HEPSGasGun1) GetConfig() Config {
	return this.config
}

func (this *HEPSGasGun1) SaveConfig(config Config) error {
	this.config = config
	return this.config.SaveLocalConfig(this.config_path)
}

func (this *HEPSGasGun1) heartbeatSnapshot() HEPSGasgun1Heartbeat {
	this.mu.RLock()
	defer this.mu.RUnlock()

	output := make(map[string]bool, len(this.heartbeat_info.Output))
	for key, value := range this.heartbeat_info.Output {
		output[key] = value
	}
	input := make(map[string]bool, len(this.heartbeat_info.Input))
	for key, value := range this.heartbeat_info.Input {
		input[key] = value
	}
	snapshot := *this.heartbeat_info
	snapshot.Output = output
	snapshot.Input = input
	return snapshot
}

func (this *HEPSGasGun1) ConnectDevice(ip string) error {
	if ip == "" {
		services.Log("连接设备失败: IP为空", this.log_path)
		return errors.New("ip is empty")
	}
	services.Log(fmt.Sprintf("连接设备: %s", ip), this.log_path)
	err := this.plc.OpenTCP(ip, 1)
	if err != nil {
		services.Log(fmt.Sprintf("连接设备失败: %v", err), this.log_path)
		return err
	}
	services.Log("设备连接成功", this.log_path)

	this.mu.Lock()
	this.heartbeat_info.Running = true
	this.mu.Unlock()
	runtime.EventsEmit(this.ctx, "heps_gasgun1_heartbeat", this.heartbeatSnapshot())

	go this.heartbeat()
	return nil
}

func (this *HEPSGasGun1) DisconnectDevice() {
	services.Log("断开设备连接", this.log_path)
	this.plc.Close()
	this.mu.Lock()
	this.heartbeat_info.Running = false
	this.mu.Unlock()
	runtime.EventsEmit(this.ctx, "heps_gasgun1_heartbeat", this.heartbeatSnapshot())
}

func (this *HEPSGasGun1) heartbeat() {
	for {
		this.mu.RLock()
		running := this.heartbeat_info.Running
		this.mu.RUnlock()
		if !running {
			return
		}
		//距离开始时间的时长HH:mm:ss
		d := int(time.Since(this.start_time).Seconds())
		this.mu.Lock()
		this.heartbeat_info.Time = fmt.Sprintf("%02d:%02d:%02d", d/3600, d%3600/60, d%60)
		this.mu.Unlock()

		// 看门狗：维持PLC心跳
		this.plc.WriteCoil(XinJie.M(this.config.DogAddr), true)

		// 输入气压
		if v, err := this.plc.ReadRegister(XinJie.D(this.config.InputPressure), XinJie.Float32, false); err == nil {
			this.mu.Lock()
			this.heartbeat_info.InputPressure = v.(float32)
			this.mu.Unlock()
		}

		// 压力（带一阶低通滤波）
		if v, err := this.plc.ReadRegister(XinJie.D(this.config.CylinderPressure), XinJie.Float32, false); err == nil {
			this.mu.Lock()
			this.heartbeat_info.CylinderPressure = v.(float32)
			this.mu.Unlock()
		}

		// 靶室真空度
		if v, err := this.plc.ReadRegister(XinJie.D(this.config.TargetVacuumDegree), XinJie.Float32, false); err == nil {
			this.mu.Lock()
			this.heartbeat_info.TargetVacuumDegree = v.(float32)
			this.mu.Unlock()
		}

		// 尾部真空度
		if v, err := this.plc.ReadRegister(XinJie.D(this.config.TailVacuumDegree), XinJie.Float32, false); err == nil {
			this.mu.Lock()
			this.heartbeat_info.TailVacuumDegree = v.(float32)
			this.mu.Unlock()
		}

		// Y线圈状态
		Y, err := this.plc.ReadCoils(XinJie.Y(0), 16, false)
		if err == nil && len(Y) >= 16 {
			aligned := make([]bool, 18)
			copy(aligned[0:8], Y[0:8])
			copy(aligned[10:18], Y[8:16])

			this.mu.Lock()
			this.heartbeat_info.Output["Fire"] = aligned[this.config.Fire]
			this.heartbeat_info.Output["Inlet"] = aligned[this.config.Inlet]
			this.heartbeat_info.Output["Outlet"] = aligned[this.config.Outlet]

			this.heartbeat_info.Output["TargetVacuumRelease"] = aligned[this.config.TargetVacuumRelease]
			this.heartbeat_info.Output["TailVacuumPump"] = aligned[this.config.TailVacuumPump]
			this.heartbeat_info.Output["TarVacuumPump"] = aligned[this.config.TarVacuumPump]
			this.heartbeat_info.Output["SystemDecompress"] = aligned[this.config.SystemDecompress]
			this.mu.Unlock()
		}

		// X输入 - 外触发信号
		if X, err := this.plc.ReadCoil(XinJie.X(0), true); err == nil {
			this.mu.Lock()
			this.heartbeat_info.Input["FireTrigger"] = X
			this.mu.Unlock()
		}

		runtime.EventsEmit(this.ctx, "heps_gasgun1_heartbeat", this.heartbeatSnapshot())

		time.Sleep(100 * time.Millisecond)
	}
}

// 抽靶室真空开关
func (g *HEPSGasGun1) TargetVacuumSwitch(open bool) error {
	services.Log(fmt.Sprintf("设置靶室真空泵: %t, 地址: Y%d", open, g.config.TarVacuumPump), g.log_path)
	if err := g.plc.WriteCoil(XinJie.Y(g.config.TarVacuumPump), open); err != nil {
		services.Log(fmt.Sprintf("设置靶室真空泵失败: %v", err), g.log_path)
		return err
	}
	return nil
}

// 抽尾部真空开关
func (g *HEPSGasGun1) TailVacuumSwitch(open bool) error {
	services.Log(fmt.Sprintf("设置尾部真空: %t, 释放阀地址: Y%d, 真空泵地址: Y%d", open,
		g.config.TargetVacuumRelease, g.config.TailVacuumPump), g.log_path)

	if err := g.plc.WriteCoil(XinJie.Y(g.config.TargetVacuumRelease), open); err != nil {
		services.Log(fmt.Sprintf("设置真空释放阀失败: %v", err), g.log_path)
		return err
	}

	time.Sleep(1000 * time.Millisecond)

	if err := g.plc.WriteCoil(XinJie.Y(g.config.TailVacuumPump), open); err != nil {
		services.Log(fmt.Sprintf("设置尾部真空泵失败: %v", err), g.log_path)
		return err
	}
	return nil
}

// 进气开关
func (g *HEPSGasGun1) InletSwitch(open bool) error {
	services.Log(fmt.Sprintf("设置进气阀: %t, 地址: Y%d", open, g.config.Inlet), g.log_path)

	if err := g.plc.WriteCoil(XinJie.Y(g.config.Inlet), open); err != nil {
		services.Log(fmt.Sprintf("设置进气阀失败: %v", err), g.log_path)
		return err
	}
	return nil
}

// 排气开关
func (g *HEPSGasGun1) ExhaustSwitch(open bool) error {
	services.Log(fmt.Sprintf("设置排气阀: %t, 地址: Y%d", open, g.config.Outlet), g.log_path)

	if err := g.plc.WriteCoil(XinJie.Y(g.config.Outlet), open); err != nil {
		services.Log(fmt.Sprintf("设置排气阀失败: %v", err), g.log_path)
		return err
	}
	return nil
}

// 系统释放阀开关
func (g *HEPSGasGun1) SystemReleaseSwitch(open bool) error {
	services.Log(fmt.Sprintf("设置系统释放阀: %t, 地址: Y%d", open, g.config.SystemDecompress), g.log_path)

	if err := g.plc.WriteCoil(XinJie.Y(g.config.SystemDecompress), open); err != nil {
		services.Log(fmt.Sprintf("设置系统释放阀失败: %v", err), g.log_path)
		return err
	}
	return nil
}

// 自动设置气室压力（通过开关进气阀与排气阀）
func (g *HEPSGasGun1) AutoCylinderPressure(target float32) {
	services.Log(fmt.Sprintf("自动设置气室压力: %f", target), g.log_path)
	g.AutoPressureFlag = true

	go func(target float32) {
		for g.AutoPressureFlag {
			g.mu.RLock()
			inlet := g.heartbeat_info.Output["Inlet"]
			outlet := g.heartbeat_info.Output["Outlet"]
			pressure := g.heartbeat_info.CylinderPressure
			g.mu.RUnlock()

			if pressure < target-0.05 {
				// 需要增压：打开进气阀，关闭排气阀
				if !inlet {
					g.InletSwitch(true)
				}
				if outlet {
					g.ExhaustSwitch(false)
				}
			} else if pressure > target+0.05 {
				// 需要减压：打开排气阀，关闭进气阀
				if !outlet {
					g.ExhaustSwitch(true)
				}
				if inlet {
					g.InletSwitch(false)
				}
			} else {
				// 达到目标压力：关闭两个阀门
				if inlet {
					g.InletSwitch(false)
				}
				if outlet {
					g.ExhaustSwitch(false)
				}
			}

			time.Sleep(100 * time.Millisecond)
		}
	}(target)
}

func (g *HEPSGasGun1) StopAutoCylinderPressure() {
	services.Log("停止自动气室压力控制", g.log_path)
	g.AutoPressureFlag = false
	// 关闭进气阀与排气阀
	g.InletSwitch(false)
	g.ExhaustSwitch(false)
}

// 发射逻辑
func (g *HEPSGasGun1) FireSwitch(ms int) error {
	services.Log(fmt.Sprintf("准备发射: 持续时间 %dms, 地址: Y%d", ms, g.config.Fire), g.log_path)

	// 发射前，先关闭抽靶室真空泵和抽尾部真空泵
	if err := g.TargetVacuumSwitch(false); err != nil {
		services.Log(fmt.Sprintf("发射准备失败，关闭靶室真空泵失败: %v", err), g.log_path)
		return err
	}
	time.Sleep(1000 * time.Millisecond)

	if err := g.TailVacuumSwitch(false); err != nil {
		services.Log(fmt.Sprintf("发射准备失败，关闭尾部真空泵失败: %v", err), g.log_path)
		return err
	}
	time.Sleep(1000 * time.Millisecond)

	err := g.plc.WriteCoil(XinJie.Y(g.config.Fire), true)
	if err != nil {
		services.Log(fmt.Sprintf("发射失败: %v", err), g.log_path)
		return err
	}
	services.Log("发射阀已打开", g.log_path)

	go func(delayMs int) {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
		err := g.plc.WriteCoil(XinJie.Y(g.config.Fire), false)
		if err != nil {
			services.Log(fmt.Sprintf("发射后自动关闭失败: %v", err), g.log_path)
			fmt.Printf("[报警] 发射后自动关闭失败: %v\n", err)
			return
		}
		services.Log("发射阀已自动关闭", g.log_path)
	}(ms)

	return nil
}

// 发射准备：关闭抽靶室真空泵和抽尾部真空（外触发模式下等待外部信号）
func (g *HEPSGasGun1) PrepareFire() error {
	services.Log("准备发射", g.log_path)

	if err := g.TargetVacuumSwitch(false); err != nil {
		services.Log(fmt.Sprintf("发射准备失败，关闭靶室真空泵失败: %v", err), g.log_path)
		return err
	}
	time.Sleep(1000 * time.Millisecond)

	if err := g.TailVacuumSwitch(false); err != nil {
		services.Log(fmt.Sprintf("发射准备失败，关闭尾部真空失败: %v", err), g.log_path)
		return err
	}

	services.Log("发射准备完成", g.log_path)
	return nil
}

// 外触发发射：直接打开发射阀 ms 毫秒后自动关闭（不含发射准备逻辑）
func (g *HEPSGasGun1) ExternalFire(ms int) error {
	services.Log(fmt.Sprintf("外触发发射: 持续时间 %dms, 地址: Y%d", ms, g.config.Fire), g.log_path)

	if err := g.plc.WriteCoil(XinJie.Y(g.config.Fire), true); err != nil {
		services.Log(fmt.Sprintf("外触发发射失败: %v", err), g.log_path)
		return err
	}

	go func(delayMs int) {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
		if err := g.plc.WriteCoil(XinJie.Y(g.config.Fire), false); err != nil {
			services.Log(fmt.Sprintf("外触发发射后自动关闭失败: %v", err), g.log_path)
			return
		}
		services.Log("外触发发射阀已自动关闭", g.log_path)
	}(ms)

	return nil
}

// ExportAlarmHistory 打开保存对话框，将历史报警消息文本写入用户选择的位置
func (this *HEPSGasGun1) ExportAlarmHistory(content string) error {
	filename, err := runtime.SaveFileDialog(this.ctx, runtime.SaveDialogOptions{
		Title:           "导出历史报警消息",
		DefaultFilename: fmt.Sprintf("报警消息_%s.txt", time.Now().Format("20060102150405")),
		Filters: []runtime.FileFilter{
			{DisplayName: "文本文件 (*.txt)", Pattern: "*.txt"},
		},
	})
	if err != nil {
		fmt.Printf("[HEPSGasGun1] 打开导出对话框失败: %v\n", err)
		return err
	}
	if filename == "" {
		return fmt.Errorf("已取消导出")
	}
	// 写入UTF-8 BOM，便于记事本/Excel正确识别中文
	if err := os.WriteFile(filename, []byte("\uFEFF"+content), 0644); err != nil {
		fmt.Printf("[HEPSGasGun1] 写入导出文件失败: %v\n", err)
		return err
	}
	fmt.Printf("[HEPSGasGun1] 历史报警消息已导出至 %s\n", filename)
	return nil
}
