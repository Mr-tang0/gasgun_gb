package xinjiegasgun2

// 监控数据结构体
type GasGun2Heartbeat struct {
	Running            bool    `json:"Running"`            //是否连接
	InputPressure      float32 `json:"InputPressure"`      //输入压力
	CylinderPressure   float32 `json:"CylinderPressure"`   //气瓶压力（一级气室）
	PumpTubePressure   float32 `json:"PumpTubePressure"`   //泵管压力（二级气室）
	PumpTubePressureHi float32 `json:"PumpTubePressureHi"` //泵管压力（高精度）

	TargetVacuumDegree float32 `json:"TargetVacuumDegree"` //靶室真空度
	TailVacuumDegree   float32 `json:"TailVacuumDegree"`   //尾部真空度
}

// 开关地址配置
type SWITCH struct {
	Pressurize uint16 `json:"Pressurize"` //增压阀
	Decompress uint16 `json:"Decompress"` //减压阀
	FireSwitch uint16 `json:"FireSwitch"` //发射阀

	PumpTubePressurize uint16 `json:"PumpTubePressurize"` //泵管增压阀
	PumpTubeDecompress uint16 `json:"PumpTubeDecompress"` //泵管减压阀

	PumpTubeVacuum    uint16 `json:"PumpTubeVacuum"`    //抽泵管真空阀
	TargetVacuum      uint16 `json:"TargetVacuum"`      //靶室真空阀
	TailVacuumProtect uint16 `json:"TailVacuumProtect"` //尾真空保护阀（尾部真空阀）
	PumpTubeProtect   uint16 `json:"PumpTubeProtect"`   //泵管保护阀

	SystemDecompress uint16 `json:"SystemDecompress"` //系统减压阀
	TargetVacuumPump uint16 `json:"TargetVacuumPump"` //靶室真空泵
	TailVacuumPump   uint16 `json:"TailVacuumPump"`   //尾真空泵
}

// 数据地址配置
type Data struct {
	InputPressure      uint16 `json:"InputPressure"`      // 输入压力地址
	CylinderPressure   uint16 `json:"CylinderPressure"`   // 气瓶压力地址
	PumpTubePressure   uint16 `json:"PumpTubePressure"`   // 泵管压力地址
	PumpTubePressureHi uint16 `json:"PumpTubePressureHi"` // 泵管压力高精度地址
	TargetVacuumDegree uint16 `json:"TargetVacuumDegree"` // 靶室真空度地址
	TailVacuumDegree   uint16 `json:"TailVacuumDegree"`   // 尾部真空度地址
}

// GasGun2Config 配置结构体
type Config struct {
	IP     string `json:"ip"`       // PLC IP地址
	Switch SWITCH `json:"switches"` // 各阀门的Modbus点位地址
	Data   Data   `json:"data"`     // 监控数据地址配置
}

func NewConfig() Config {
	return Config{
		IP: "192.168.6.6",
		Switch: SWITCH{
			Pressurize:         0,
			Decompress:         1,
			PumpTubePressurize: 4,
			PumpTubeDecompress: 5,
			PumpTubeVacuum:     7,
			TargetVacuum:       13,
			TailVacuumProtect:  10,
			PumpTubeProtect:    6,
			FireSwitch:         2,
			SystemDecompress:   3,
			TargetVacuumPump:   11,
			TailVacuumPump:     12,
		},
		Data: Data{
			InputPressure:      56,
			CylinderPressure:   54,
			PumpTubePressure:   50,
			PumpTubePressureHi: 52,
			TargetVacuumDegree: 58,
			TailVacuumDegree:   60,
		},
	}
}

func (c *Config) LoadLocalConfig() error {
	return nil
}

func (c *Config) SaveLocalConfig() error {
	return nil
}
