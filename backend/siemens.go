package backend

import (
	"context"
	"sync"
	"time"

	"github.com/goburrow/modbus"
)

type Device struct {
	Name   string
	IP     string
	Opened bool
}

type Siemens struct {
	ctx           context.Context
	SiemensDevice Device

	handler *modbus.TCPClientHandler // Changed from RTU to TCP
	client  modbus.Client
	mu      sync.Mutex
}

func NewSiemens(ctx context.Context) *Siemens {
	return &Siemens{}
}

func (s *Siemens) startup() {

}

func (s *Siemens) EnumDevices() []Device {

	return []Device{{Name: "Siemens", IP: "192.168.1.1"}}
}

func (s *Siemens) Connect(device Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.handler != nil {
		s.handler.Close()
	}

	s.handler = modbus.NewTCPClientHandler(device.IP)
	s.handler.Timeout = 2 * time.Second
	s.handler.SlaveId = 1

	err := s.handler.Connect()
	if err != nil {
		return err
	}

	s.client = modbus.NewClient(s.handler)
	return nil
}

func (s *Siemens) Disconnect() error {
	return nil
}
