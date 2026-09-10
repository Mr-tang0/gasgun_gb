package hepsgasgun1

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	XinJie "github.com/Mr-tang0/PIMSGoMod/protocol/XinJie"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type HEPSGasGun1 struct {
	plc             *XinJie.XinjieClient
	ctx             context.Context
	heartbeat_info  *HEPSGasgunHeartbeat
	auto_pressurize bool

	start_time  time.Time
	config      Config
	config_path string
}

func NewHEPSGasGun1() *HEPSGasGun1 {
	return &HEPSGasGun1{
		plc:    XinJie.NewXinjieClient("XD"),
		config: NewConfig(),
		heartbeat_info: &HEPSGasgunHeartbeat{
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

func (this *HEPSGasGun1) Startup(ctx context.Context) {
	this.ctx = ctx
	this.start_time = time.Now()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		this.config_path = "config.json"
	} else {
		this.config_path = filepath.Join(homeDir, "PIMS", "GASGUN", "xinjie_heps_gasgun_1", "config.json")
	}
	this.config.LoadLocalConfig(this.config_path)
}

func (this *HEPSGasGun1) GetConfig() Config {
	return this.config
}

func (this *HEPSGasGun1) SaveConfig(config Config) error {
	this.config = config
	return this.config.SaveLocalConfig(this.config_path)
}

func (this *HEPSGasGun1) ConnectDevice(ip string) error {
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

func (this *HEPSGasGun1) DisconnectDevice() {
	this.plc.Close()
	this.heartbeat_info.Running = false
	runtime.EventsEmit(this.ctx, "heartbeat", this.heartbeat_info)
}

func (this *HEPSGasGun1) heartbeat() {
	lastDog := time.Now()
	dogState := false
	for {
		if !this.heartbeat_info.Running {
			return
		}
		//距离开始时间的时长HH:mm:ss
		d := int(time.Since(this.start_time).Seconds())
		this.heartbeat_info.Time = fmt.Sprintf("%02d:%02d:%02d", d/3600, d%3600/60, d%60)

		vacuumFloat, err := this.plc.ReadRegister(XinJie.D(this.config.VacuumFloatAddr), XinJie.Float32, false)
		if err != nil {
			this.AddMessageToFrontend(err.Error())
			this.DisconnectDevice()
			return
		}
		this.heartbeat_info.Vacuum = vacuumFloat.(float32)

		pressureFloat, err := this.plc.ReadRegister(XinJie.D(this.config.PressureFloatAddr), XinJie.Float32, false)
		if err != nil {
			this.AddMessageToFrontend(err.Error())
			this.DisconnectDevice()
			return
		}
		this.heartbeat_info.Pressure = pressureFloat.(float32)

		Y, err := this.plc.ReadCoils(XinJie.Y(0), 8, false)
		if err != nil {
			this.AddMessageToFrontend(err.Error())
			this.DisconnectDevice()
			return
		}
		// fmt.Printf("Y: %v\n", Y)

		// Y的长度必须大于等于8
		if len(Y) < 8 {
			this.AddMessageToFrontend("Y的长度必须大于等于8")
			return
		}

		this.heartbeat_info.OUTPUT["Fire"] = Y[this.config.FireAddr]
		this.heartbeat_info.OUTPUT["Inlet"] = Y[this.config.InletAddr]
		this.heartbeat_info.OUTPUT["Outlet"] = Y[this.config.OutletAddr]
		this.heartbeat_info.OUTPUT["VacuumRealse"] = Y[this.config.VacuumRealseAddr]
		// this.heartbeat_info.OUTPUT["PressureOpen"] = Y[this.config.PressureOpenAddr]
		// this.heartbeat_info.OUTPUT["PressureClose"] = Y[this.config.PressureCloseAddr]
		this.heartbeat_info.OUTPUT["TailVacuumPump"] = Y[this.config.TailVacuumPumpAddr]
		this.heartbeat_info.OUTPUT["TarVacuumPump"] = Y[this.config.TarVacuumPumpAddr]
		// fmt.Printf("OUTPUT: %v\n", this.heartbeat_info.OUTPUT)

		runtime.EventsEmit(this.ctx, "heartbeat", this.heartbeat_info)

		// 每隔2s向看门狗线圈写一次，维持PLC心跳
		if time.Since(lastDog) >= 2*time.Second {
			dogState = !dogState
			if err := this.plc.WriteCoil(XinJie.M(this.config.DogAddr), dogState); err == nil {
				lastDog = time.Now()
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (g *HEPSGasGun1) AddMessageToFrontend(str string) {
	runtime.EventsEmit(g.ctx, "message", str)
}

// 进气开关
func (g *HEPSGasGun1) InletSwitch(open bool) error {
	return g.plc.WriteCoil(XinJie.Y(g.config.InletAddr), open)
}

// 排气开关
func (g *HEPSGasGun1) ExhaustSwitch(open bool) error {
	return g.plc.WriteCoil(XinJie.Y(g.config.OutletAddr), open)
}

// 发射逻辑
func (g *HEPSGasGun1) FireSwitch(ms int) error {

	// 发射前，先关闭抽靶室真空泵和抽尾部真空泵
	g.VacuumSwitch(false)
	time.Sleep(1000 * time.Millisecond)
	g.TailVacuumSwitch(false)
	time.Sleep(1000 * time.Millisecond)
	g.PressureSwitch(false)
	time.Sleep(1000 * time.Millisecond)

	err := g.plc.WriteCoil(XinJie.Y(g.config.FireAddr), true)
	if err != nil {
		return err
	}

	go func(delayMs int) {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
		err := g.plc.WriteCoil(XinJie.Y(g.config.FireAddr), false)
		if err != nil {
			fmt.Printf("[报警] 发射后自动关闭失败: %v\n", err)
		}
	}(ms)

	return nil
}

// 抽靶室真空开关
func (g *HEPSGasGun1) VacuumSwitch(open bool) error {
	return g.plc.WriteCoil(XinJie.Y(g.config.TarVacuumPumpAddr), open)
}

// 抽尾部真空开关
func (g *HEPSGasGun1) TailVacuumSwitch(open bool) error {
	g.plc.WriteCoil(XinJie.Y(g.config.VacuumRealseAddr), open)
	time.Sleep(1000 * time.Millisecond)
	return g.plc.WriteCoil(XinJie.Y(g.config.TailVacuumPumpAddr), open)
}

// 增压泵开关
func (g *HEPSGasGun1) PressureSwitch(open bool) error {
	if open {
		err := g.plc.WriteCoil(XinJie.Y(g.config.PressureCloseAddr), true)
		if err != nil {
			return err
		}
		err = g.plc.WriteCoil(XinJie.Y(g.config.PressureOpenAddr), true)
		if err != nil {
			return err
		}

		time.Sleep(500 * time.Millisecond)
		err = g.plc.WriteCoil(XinJie.Y(g.config.PressureOpenAddr), false)
		if err != nil {
			return err
		}
		g.heartbeat_info.OUTPUT["PressureOpen"] = true
		g.heartbeat_info.OUTPUT["PressureClose"] = false
		return nil
	} else {
		err := g.plc.WriteCoil(XinJie.Y(g.config.PressureOpenAddr), false)
		if err != nil {
			return err
		}
		err = g.plc.WriteCoil(XinJie.Y(g.config.PressureCloseAddr), false)
		if err != nil {
			return err
		}
		g.heartbeat_info.OUTPUT["PressureOpen"] = false
		g.heartbeat_info.OUTPUT["PressureClose"] = true
		return nil
	}
}

// 自动增压,通过比较this.heartbeat_info.Pressure与target的值，来自动控制InletSwitch与ExhaustSwitch的开关，用以保持压力在target值附近（+/-0.2）
func (g *HEPSGasGun1) AutoPressurize(target float32) {
	g.auto_pressurize = true

	go func(target float32) {
		for {
			if !g.auto_pressurize {
				break
			}

			if g.heartbeat_info.Pressure > target+0.1 {
				//1、当前气压大于目标值+0.1，排气开关未打开则打开排气开关， 进气开关未关闭则关闭进气开关
				if !g.heartbeat_info.OUTPUT["Outlet"] {
					g.InletSwitch(true)
				}
				if g.heartbeat_info.OUTPUT["Inlet"] {
					g.ExhaustSwitch(false)
				}
			} else if g.heartbeat_info.Pressure < target-0.1 {
				//2、当前气压小于目标值-0.1，进气开关未打开则打开进气开关， 排气开关未关闭则关闭排气开关
				if !g.heartbeat_info.OUTPUT["Inlet"] {
					g.InletSwitch(true)
				}
				if g.heartbeat_info.OUTPUT["Outlet"] {
					g.ExhaustSwitch(false)
				}
			} else {
				//3、当前气压在目标值+0.1到目标值-0.1之间，进气开关关闭，排气开关关闭
				if g.heartbeat_info.OUTPUT["Inlet"] {
					g.InletSwitch(false)
				}
				if g.heartbeat_info.OUTPUT["Outlet"] {
					g.ExhaustSwitch(false)
				}
			}

			time.Sleep(100 * time.Millisecond)
		}
	}(target)
}

func (g *HEPSGasGun1) StopAutoPressurize() {
	g.auto_pressurize = false

	time.Sleep(500 * time.Millisecond)

	g.InletSwitch(false)
	g.ExhaustSwitch(false)
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
