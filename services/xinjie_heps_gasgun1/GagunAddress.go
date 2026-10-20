package hepsgasgun1

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type HEPSGasgun1Heartbeat struct {
	Running bool   `json:"Running"` //是否连接
	Time    string `json:"Time"`
	Alarm   string `json:"Alarm"`

	InputPressure      float32 `json:"InputPressure"`      //输入压力
	CylinderPressure   float32 `json:"CylinderPressure"`   //气瓶压力
	TargetVacuumDegree float32 `json:"TargetVacuumDegree"` //靶室真空度
	TailVacuumDegree   float32 `json:"TailVacuumDegree"`   //尾部真空度

	Output map[string]bool `json:"Output"` //各阀门状态
	Input  map[string]bool `json:"Input"`  //各阀门输入状态
}

// 地址配置
type Config struct {
	// Y地址
	Inlet  uint16 `json:"Inlet"`  // 进气开关地址
	Outlet uint16 `json:"Outlet"` // 出气开关地址
	Fire   uint16 `json:"Fire"`   // 发射开关地址

	TailVacuumPump      uint16 `json:"TailVacuumPump"`      // 尾部真空泵地址
	TargetVacuumRelease uint16 `json:"TargetVacuumRelease"` // 靶室真空阀地址
	TarVacuumPump       uint16 `json:"TarVacuumPump"`       // 靶室真空泵地址
	SystemDecompress    uint16 `json:"SystemDecompress"`    // 系统减压阀

	// D地址
	InputPressure      uint16 `json:"InputPressure"`      // 输入压力地址
	CylinderPressure   uint16 `json:"CylinderPressure"`   // 气瓶压力地址
	TargetVacuumDegree uint16 `json:"TargetVacuumDegree"` // 靶室真空度地址
	TailVacuumDegree   uint16 `json:"TailVacuumDegree"`   // 尾部真空度地址

	// M地址
	DogAddr uint16 `json:"DogAddr"` // 看门狗地址

	//X地址
	FireTrigger uint16 `json:"FireTrigger"` // 发射触发地址
}

func NewConfig() Config {
	return Config{
		Inlet:  0,
		Outlet: 1,
		Fire:   2,

		TailVacuumPump:      12,
		TargetVacuumRelease: 13,
		TarVacuumPump:       11,
		SystemDecompress:    3,

		InputPressure:      56,
		CylinderPressure:   54,
		TargetVacuumDegree: 58,
		TailVacuumDegree:   60,

		DogAddr: 0,
	}
}

func (c *Config) LoadLocalConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, c)
}

func (c *Config) SaveLocalConfig(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
