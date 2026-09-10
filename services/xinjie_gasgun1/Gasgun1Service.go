package xinjiegasgun1

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

type XinjieGasGun1 struct {
	plc             *XinJie.XinjieClient
	ctx             context.Context
	heartbeat_info  *GasgunHeartbeat
	auto_pressurize bool

	start_time time.Time

	config      Config
	config_path string
	log_path    string

	mu sync.RWMutex // 保护心跳状态和自动增压标志的并发访问
}

func NewXinjieGasGun1() *XinjieGasGun1 {
	return &XinjieGasGun1{
		plc:    XinJie.NewXinjieClient("XD"),
		config: NewConfig(),
		heartbeat_info: &GasgunHeartbeat{
			Running:  false,
			Alarm:    "",
			Vacuum:   100000,
			Pressure: 0,
			OUTPUT: map[string]bool{
				"Fire":           false,
				"Inlet":          false,
				"Outlet":         false,
				"VacuumRealse":   false,
				"PressureOpen":   false,
				"PressureClose":  false,
				"TailVacuumPump": false,
				"TarVacuumPump":  false,
			},
		},
	}
}

func (this *XinjieGasGun1) Startup(ctx context.Context) {
	this.ctx = ctx
	this.start_time = time.Now()
	this.log_path = services.GetLogPath("NIMTE", "GASGUN", "xinjie_gasgun_1")
	this.config_path = services.GetConfigPath("NIMTE", "GASGUN", "xinjie_gasgun_1")

	fmt.Println(this.config_path, this.log_path)
	this.config.LoadLocalConfig(this.config_path)
}

func (this *XinjieGasGun1) GetConfig() Config {
	return this.config
}

func (this *XinjieGasGun1) SaveConfig(config Config) error {
	this.config = config
	return this.config.SaveLocalConfig(this.config_path)
}

func (this *XinjieGasGun1) heartbeatSnapshot() GasgunHeartbeat {
	this.mu.RLock()
	defer this.mu.RUnlock()

	output := make(map[string]bool, len(this.heartbeat_info.OUTPUT))
	for key, value := range this.heartbeat_info.OUTPUT {
		output[key] = value
	}
	snapshot := *this.heartbeat_info
	snapshot.OUTPUT = output
	return snapshot
}

func (this *XinjieGasGun1) ConnectDevice(ip string) error {
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
	runtime.EventsEmit(this.ctx, "heartbeat", this.heartbeatSnapshot())

	go this.heartbeat()
	return nil
}

func (this *XinjieGasGun1) DisconnectDevice() {
	services.Log("断开设备连接", this.log_path)
	this.plc.Close()
	this.mu.Lock()
	this.heartbeat_info.Running = false
	this.mu.Unlock()
	runtime.EventsEmit(this.ctx, "heartbeat", this.heartbeatSnapshot())
}

func (this *XinjieGasGun1) heartbeat() {
	lastDog := time.Now()
	dogState := false
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

		vacuumFloat, err := this.plc.ReadRegister(XinJie.D(this.config.VacuumFloatAddr), XinJie.Float32, false)
		if err != nil {
			this.AddMessageToFrontend(err.Error())
			this.DisconnectDevice()
			return
		}
		this.mu.Lock()
		this.heartbeat_info.Vacuum = vacuumFloat.(float32)
		this.mu.Unlock()

		pressureFloat, err := this.plc.ReadRegister(XinJie.D(this.config.PressureFloatAddr), XinJie.Float32, false)

		if err != nil {
			this.AddMessageToFrontend(err.Error())
			this.DisconnectDevice()
			return
		}
		this.mu.Lock()
		const alpha = 0.3
		if this.heartbeat_info.Pressure == 0 {
			this.heartbeat_info.Pressure = pressureFloat.(float32)
		} else {
			this.heartbeat_info.Pressure += alpha * (pressureFloat.(float32) - this.heartbeat_info.Pressure)
		}
		this.mu.Unlock()

		Y, err := this.plc.ReadCoils(XinJie.Y(0), 8, false)
		if err != nil {
			this.AddMessageToFrontend(err.Error())
			this.DisconnectDevice()
			return
		}

		// Y的长度必须大于等于8
		if len(Y) < 8 {
			this.AddMessageToFrontend("Y的长度必须大于等于8")
			return
		}

		this.mu.Lock()
		this.heartbeat_info.OUTPUT["Fire"] = Y[this.config.FireAddr]
		this.heartbeat_info.OUTPUT["Inlet"] = Y[this.config.InletAddr]
		this.heartbeat_info.OUTPUT["Outlet"] = Y[this.config.OutletAddr]
		this.heartbeat_info.OUTPUT["VacuumRealse"] = Y[this.config.VacuumRealseAddr]
		this.heartbeat_info.OUTPUT["TailVacuumPump"] = Y[this.config.TailVacuumPumpAddr]
		this.heartbeat_info.OUTPUT["TarVacuumPump"] = Y[this.config.TarVacuumPumpAddr]
		this.mu.Unlock()

		runtime.EventsEmit(this.ctx, "heartbeat", this.heartbeatSnapshot())

		// 每隔2s向看门狗线圈写一次，维持PLC心跳
		if time.Since(lastDog) >= 2*time.Second {
			dogState = !dogState
			if err := this.plc.WriteCoil(XinJie.M(this.config.DogAddr), dogState); err == nil {
				lastDog = time.Now()
			} else {
				services.Log(fmt.Sprintf("写入看门狗失败: %v", err), this.log_path)
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (g *XinjieGasGun1) AddMessageToFrontend(str string) {
	runtime.EventsEmit(g.ctx, "message", str)
}

// 进气开关
func (g *XinjieGasGun1) InletSwitch(open bool) error {
	services.Log(fmt.Sprintf("设置进气阀: %t, 地址: Y%d", open, g.config.InletAddr), g.log_path)
	if err := g.plc.WriteCoil(XinJie.Y(g.config.InletAddr), open); err != nil {
		services.Log(fmt.Sprintf("设置进气阀失败: %v", err), g.log_path)
		return err
	}
	return nil
}

// 排气开关
func (g *XinjieGasGun1) ExhaustSwitch(open bool) error {
	services.Log(fmt.Sprintf("设置排气阀: %t, 地址: Y%d", open, g.config.OutletAddr), g.log_path)
	if err := g.plc.WriteCoil(XinJie.Y(g.config.OutletAddr), open); err != nil {
		services.Log(fmt.Sprintf("设置排气阀失败: %v", err), g.log_path)
		return err
	}
	return nil
}

// 发射逻辑
func (g *XinjieGasGun1) FireSwitch(ms int) error {
	services.Log(fmt.Sprintf("准备发射: 持续时间 %dms, 地址: Y%d", ms, g.config.FireAddr), g.log_path)

	// 发射前，先关闭抽靶室真空泵和抽尾部真空泵
	if err := g.VacuumSwitch(false); err != nil {
		services.Log(fmt.Sprintf("发射准备失败，关闭靶室真空泵失败: %v", err), g.log_path)
		return err
	}
	time.Sleep(1000 * time.Millisecond)
	if err := g.TailVacuumSwitch(false); err != nil {
		services.Log(fmt.Sprintf("发射准备失败，关闭尾部真空泵失败: %v", err), g.log_path)
		return err
	}
	time.Sleep(1000 * time.Millisecond)
	if err := g.PressureSwitch(false); err != nil {
		services.Log(fmt.Sprintf("发射准备失败，关闭增压泵失败: %v", err), g.log_path)
		return err
	}
	time.Sleep(1000 * time.Millisecond)

	err := g.plc.WriteCoil(XinJie.Y(g.config.FireAddr), true)
	if err != nil {
		services.Log(fmt.Sprintf("发射失败: %v", err), g.log_path)
		return err
	}
	services.Log("发射阀已打开", g.log_path)

	go func(delayMs int) {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
		err := g.plc.WriteCoil(XinJie.Y(g.config.FireAddr), false)
		if err != nil {
			services.Log(fmt.Sprintf("发射后自动关闭失败: %v", err), g.log_path)
			fmt.Printf("[报警] 发射后自动关闭失败: %v\n", err)
			return
		}
		services.Log("发射阀已自动关闭", g.log_path)
	}(ms)

	return nil
}

// 抽靶室真空开关
func (g *XinjieGasGun1) VacuumSwitch(open bool) error {
	services.Log(fmt.Sprintf("设置靶室真空泵: %t, 地址: Y%d", open, g.config.TarVacuumPumpAddr), g.log_path)
	if err := g.plc.WriteCoil(XinJie.Y(g.config.TarVacuumPumpAddr), open); err != nil {
		services.Log(fmt.Sprintf("设置靶室真空泵失败: %v", err), g.log_path)
		return err
	}
	return nil
}

// 抽尾部真空开关
func (g *XinjieGasGun1) TailVacuumSwitch(open bool) error {
	services.Log(fmt.Sprintf("设置尾部真空: %t, 释放阀地址: Y%d, 真空泵地址: Y%d", open, g.config.VacuumRealseAddr, g.config.TailVacuumPumpAddr), g.log_path)
	if err := g.plc.WriteCoil(XinJie.Y(g.config.VacuumRealseAddr), open); err != nil {
		services.Log(fmt.Sprintf("设置真空释放阀失败: %v", err), g.log_path)
		return err
	}
	time.Sleep(1000 * time.Millisecond)
	if err := g.plc.WriteCoil(XinJie.Y(g.config.TailVacuumPumpAddr), open); err != nil {
		services.Log(fmt.Sprintf("设置尾部真空泵失败: %v", err), g.log_path)
		return err
	}
	return nil
}

// 增压泵开关
func (g *XinjieGasGun1) PressureSwitch(open bool) error {
	services.Log(fmt.Sprintf("设置增压泵: %t, 开启地址: Y%d, 关闭地址: Y%d", open, g.config.PressureOpenAddr, g.config.PressureCloseAddr), g.log_path)
	if open {
		err := g.plc.WriteCoil(XinJie.Y(g.config.PressureCloseAddr), true)
		if err != nil {
			services.Log(fmt.Sprintf("设置增压泵关闭阀失败: %v", err), g.log_path)
			return err
		}
		err = g.plc.WriteCoil(XinJie.Y(g.config.PressureOpenAddr), true)
		if err != nil {
			services.Log(fmt.Sprintf("打开增压泵失败: %v", err), g.log_path)
			return err
		}

		time.Sleep(500 * time.Millisecond)
		err = g.plc.WriteCoil(XinJie.Y(g.config.PressureOpenAddr), false)
		if err != nil {
			services.Log(fmt.Sprintf("复位增压泵开启阀失败: %v", err), g.log_path)
			return err
		}
		g.mu.Lock()
		g.heartbeat_info.OUTPUT["PressureOpen"] = true
		g.heartbeat_info.OUTPUT["PressureClose"] = false
		g.mu.Unlock()
		services.Log("增压泵已开启", g.log_path)
		return nil
	} else {
		err := g.plc.WriteCoil(XinJie.Y(g.config.PressureOpenAddr), false)
		if err != nil {
			services.Log(fmt.Sprintf("关闭增压泵开启阀失败: %v", err), g.log_path)
			return err
		}
		err = g.plc.WriteCoil(XinJie.Y(g.config.PressureCloseAddr), false)
		if err != nil {
			services.Log(fmt.Sprintf("关闭增压泵关闭阀失败: %v", err), g.log_path)
			return err
		}
		g.mu.Lock()
		g.heartbeat_info.OUTPUT["PressureOpen"] = false
		g.heartbeat_info.OUTPUT["PressureClose"] = true
		g.mu.Unlock()
		services.Log("增压泵已关闭", g.log_path)
		return nil
	}
}

// 自动增压,通过比较this.heartbeat_info.Pressure与target的值，来自动控制InletSwitch与ExhaustSwitch的开关，用以保持压力在target值附近（+/-0.2）
func (g *XinjieGasGun1) AutoPressurize(target float32) {
	services.Log(fmt.Sprintf("启动自动增压控制: 目标压力 %.3f", target), g.log_path)
	g.mu.Lock()
	g.auto_pressurize = true
	g.mu.Unlock()

	go func(target float32) {
		for {
			g.mu.RLock()
			autoPressurize := g.auto_pressurize
			pressure := g.heartbeat_info.Pressure
			inlet := g.heartbeat_info.OUTPUT["Inlet"]
			outlet := g.heartbeat_info.OUTPUT["Outlet"]
			g.mu.RUnlock()

			if !autoPressurize {
				break
			}

			if pressure > target+0.1 {
				//1、当前气压大于目标值+0.1，排气开关未打开则打开排气开关， 进气开关未关闭则关闭进气开关
				if !outlet {
					g.InletSwitch(true)
				}
				if inlet {
					g.ExhaustSwitch(false)
				}
			} else if pressure < target-0.1 {
				//2、当前气压小于目标值-0.1，进气开关未打开则打开进气开关， 排气开关未关闭则关闭排气开关
				if !inlet {
					g.InletSwitch(true)
				}
				if outlet {
					g.ExhaustSwitch(false)
				}
			} else {
				//3、当前气压在目标值+0.1到目标值-0.1之间，进气开关关闭，排气开关关闭
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

func (g *XinjieGasGun1) StopAutoPressurize() {
	services.Log("停止自动增压控制", g.log_path)
	g.mu.Lock()
	g.auto_pressurize = false
	g.mu.Unlock()

	time.Sleep(500 * time.Millisecond)

	if err := g.InletSwitch(false); err != nil {
		services.Log(fmt.Sprintf("停止自动增压时关闭进气阀失败: %v", err), g.log_path)
	}
	if err := g.ExhaustSwitch(false); err != nil {
		services.Log(fmt.Sprintf("停止自动增压时关闭排气阀失败: %v", err), g.log_path)
	}
}

// ExportAlarmHistory 打开保存对话框，将历史报警消息文本写入用户选择的位置
func (this *XinjieGasGun1) ExportAlarmHistory(content string) error {
	filename, err := runtime.SaveFileDialog(this.ctx, runtime.SaveDialogOptions{
		Title:           "导出历史报警消息",
		DefaultFilename: fmt.Sprintf("报警消息_%s.txt", time.Now().Format("20060102150405")),
		Filters: []runtime.FileFilter{
			{DisplayName: "文本文件 (*.txt)", Pattern: "*.txt"},
		},
	})
	if err != nil {
		fmt.Printf("[DetectorService] 打开导出对话框失败: %v\n", err)
		return err
	}
	if filename == "" {
		return fmt.Errorf("已取消导出")
	}
	// 写入UTF-8 BOM，便于记事本/Excel正确识别中文
	if err := os.WriteFile(filename, []byte("\uFEFF"+content), 0644); err != nil {
		fmt.Printf("[DetectorService] 写入导出文件失败: %v\n", err)
		return err
	}
	fmt.Printf("[DetectorService] 历史报警消息已导出至 %s\n", filename)
	return nil
}
