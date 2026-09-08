package xinjiegasgun1

const (
	//Y地址
	InletAddr          = 0
	OutletAddr         = 1
	FireAddr           = 2
	VacuumRealseAddr   = 3
	PressureOpenAddr   = 4
	PressureCloseAddr  = 5
	TailVacuumPumpAddr = 6
	TarVacuumPumpAddr  = 7

	//D地址
	VacuumFloatAddr   = 0
	PressureFloatAddr = 2

	//M地址
	DogAddr = 0
)

type GasgunHeartbeat struct {
	Time     string          `json:"time"`
	Running  bool            `json:"running"`
	Alarm    string          `json:"alarm"`
	Vacuum   float32         `json:"vacuum"`
	Pressure float32         `json:"pressure"`
	OUTPUT   map[string]bool `json:"output"`
}
