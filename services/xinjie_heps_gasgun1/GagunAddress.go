package hepsgasgun1

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// 地址配置
type Config struct {
	// Y地址
	InletAddr          uint16 `json:"InletAddr"`
	OutletAddr         uint16 `json:"OutletAddr"`
	FireAddr           uint16 `json:"FireAddr"`
	VacuumRealseAddr   uint16 `json:"VacuumRealseAddr"`
	PressureOpenAddr   uint16 `json:"PressureOpenAddr"`
	PressureCloseAddr  uint16 `json:"PressureCloseAddr"`
	TailVacuumPumpAddr uint16 `json:"TailVacuumPumpAddr"`
	TarVacuumPumpAddr  uint16 `json:"TarVacuumPumpAddr"`

	// D地址
	InputPressureAddr   uint16 `json:"InputPressureAddr"`
	VacuumFloatAddr     uint16 `json:"VacuumFloatAddr"`
	PressureFloatAddr   uint16 `json:"PressureFloatAddr"`
	TailVaccumFloatAddr uint16 `json:"TailVaccumFloatAddr"`

	// M地址
	DogAddr uint16 `json:"DogAddr"`
}

func NewConfig() Config {
	return Config{
		InletAddr:           0,
		OutletAddr:          1,
		FireAddr:            2,
		VacuumRealseAddr:    10,
		PressureOpenAddr:    4,
		PressureCloseAddr:   5,
		TailVacuumPumpAddr:  12,
		TarVacuumPumpAddr:   11,
		InputPressureAddr:   56,
		VacuumFloatAddr:     58,
		PressureFloatAddr:   54,
		TailVaccumFloatAddr: 60,
		DogAddr:             0,
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

type HEPSGasgunHeartbeat struct {
	Time     string          `json:"time"`
	Running  bool            `json:"running"`
	Alarm    string          `json:"alarm"`
	Vacuum   float32         `json:"vacuum"`
	Pressure float32         `json:"pressure"`
	OUTPUT   map[string]bool `json:"output"`
}
