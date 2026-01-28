package backend

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"regexp"
	"strconv"
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
	ctx context.Context

	handler *modbus.TCPClientHandler // Changed from RTU to TCP
	client  modbus.Client
	mu      sync.Mutex
}

func NewSiemens(ctx context.Context) *Siemens {
	return &Siemens{ctx: ctx}
}

func (s *Siemens) startup() {

}

func (s *Siemens) EnumDevices() []Device {

	return []Device{{Name: "Siemens", IP: "192.168.1.1"}}
}

func (s *Siemens) Connect(IP string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.handler != nil {
		s.handler.Close()
	}

	s.handler = modbus.NewTCPClientHandler(IP + ":502")
	s.handler.Timeout = 2 * time.Second
	s.handler.SlaveId = 1

	err := s.handler.Connect()
	if err != nil {
		fmt.Println(err)
		return err
	}

	s.client = modbus.NewClient(s.handler)
	return nil
}

func (s *Siemens) Disconnect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.handler != nil {
		err := s.handler.Close()
		s.handler = nil
		s.client = nil
		fmt.Println("PLC连接已断开，句柄已销毁")
		return err
	}
	return nil
}

// parseAddress 将 "VW10", "Q0.0", "I0.1" 转换为 Modbus 偏移量
// parseAddress 增强版：支持 VB, VW, VD, Q, I
func (s *Siemens) parseAddress(addrStr string) (offset uint16, area string, err error) {
	re := regexp.MustCompile(`^([A-Z]+)(\d+)(\.?\d+)?$`)
	matches := re.FindStringSubmatch(addrStr)
	if len(matches) < 3 {
		return 0, "", fmt.Errorf("地址格式错误: %s", addrStr)
	}

	area = matches[1]
	baseAddr, _ := strconv.Atoi(matches[2])

	switch area {
	case "VB", "VW", "VD":
		// 由于 Modbus 是字寻址（2字节），所有 V 区地址都映射到 Holding Register
		// 映射公式：Modbus偏移量 = PLC字节地址 / 2
		offset = uint16(baseAddr / 2)
		// 注意：如果读 VB1，它其实在 Modbus 地址 0 的低位字节（VB1 = VW0 的低位）
	case "Q", "I":
		bitPart := 0
		if matches[3] != "" {
			subAddr, _ := strconv.Atoi(matches[3][1:])
			bitPart = subAddr
		}
		offset = uint16(baseAddr*8 + bitPart)
	default:
		return 0, "", fmt.Errorf("不支持的区域: %s", area)
	}
	return offset, area, nil
}

// ReadVW 传入 "VW0", 2 代表读 VW0, VW2
func (s *Siemens) ReadVW(addrStr string, quantity uint16) ([]uint16, error) {
	offset, area, err := s.parseAddress(addrStr)
	if err != nil || area != "VW" {
		return nil, fmt.Errorf("无效的V区地址: %v", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	bytes, err := s.client.ReadHoldingRegisters(offset, quantity)
	if err != nil {
		return nil, err
	}

	values := make([]uint16, quantity)
	for i := 0; i < int(quantity); i++ {
		values[i] = binary.BigEndian.Uint16(bytes[i*2 : i*2+2])
	}
	return values, nil
}

func (s *Siemens) WriteVW(addrStr string, value uint16) error {
	offset, area, err := s.parseAddress(addrStr)
	if err != nil || area != "VW" {
		return fmt.Errorf("无效的V区地址: %v", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	_, err = s.client.WriteSingleRegister(offset, value)
	return err
}

func (s *Siemens) ReadVD(addrStr string) (uint32, error) {
	offset, area, err := s.parseAddress(addrStr)
	if err != nil || area != "VD" {
		return 0, fmt.Errorf("无效的VD地址")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	// VD 占用 2 个寄存器
	bytes, err := s.client.ReadHoldingRegisters(offset, 2)
	if err != nil {
		return 0, err
	}
	// 西门子是大端模式 (BigEndian)
	return binary.BigEndian.Uint32(bytes), nil
}

// WriteVD 写 VD 地址
func (s *Siemens) WriteVD(addrStr string, value uint32) error {
	offset, area, err := s.parseAddress(addrStr)
	if err != nil || area != "VD" {
		return fmt.Errorf("无效的VD地址")
	}

	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, value)

	s.mu.Lock()
	defer s.mu.Unlock()
	// 使用功能码 16 写多个寄存器
	_, err = s.client.WriteMultipleRegisters(offset, 2, buf)
	return err
}

// ReadVDFloats 批量读取连续的 VD 浮点数
// addrStr: 起始地址，如 "VD100"
// count: 要读取的浮点数个数（每个浮点数占 2 个寄存器）
func (s *Siemens) ReadVDFloats(addrStr string, count uint16) ([]float32, error) {
	// 1. 解析起始地址偏移
	offset, area, err := s.parseAddress(addrStr)
	if err != nil || area != "VD" {
		return nil, fmt.Errorf("无效的VD起始地址: %s", addrStr)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 2. 每个 VD 占 2 个寄存器，总共读取 count * 2 个寄存器
	// 注意：Modbus 单次读取寄存器数量通常限制在 125 个以内
	bytes, err := s.client.ReadHoldingRegisters(offset, count*2)
	if err != nil {
		return nil, fmt.Errorf("批量读取寄存器失败: %v", err)
	}

	// 3. 循环解析字节数组
	results := make([]float32, count)
	for i := 0; i < int(count); i++ {
		// 每个浮点数占用 4 个字节（i*4）
		start := i * 4
		bits := binary.BigEndian.Uint32(bytes[start : start+4])
		results[i] = math.Float32frombits(bits)
	}

	return results, nil
}

// WriteVDFloat 向 VD 地址写入 float32 浮点数
func (s *Siemens) WriteVDFloat(addrStr string, value float32) error {
	// 1. 解析地址得到偏移量
	offset, area, err := s.parseAddress(addrStr)
	if err != nil || area != "VD" {
		return fmt.Errorf("无效的VD地址: %s", addrStr)
	}

	// 2. 关键步骤：将 float32 转换为 IEEE 754 的位模式 (uint32)
	// 这是将 3.14 转换为类似 0x4048F5C3 的二进制位
	bits := math.Float32bits(value)

	// 3. 将 uint32 放入 4 字节缓冲区 (西门子使用大端序)
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, bits)

	s.mu.Lock()
	defer s.mu.Unlock()

	// 4. 使用 Modbus 功能码 16 (Write Multiple Registers) 写入 2 个寄存器
	_, err = s.client.WriteMultipleRegisters(offset, 2, buf)
	if err != nil {
		return fmt.Errorf("写入VD失败: %v", err)
	}

	return nil
}

// ReadQ 读输出线圈，如 ReadQ("Q0.0", 8) 读 Q0.0-Q0.7
func (s *Siemens) ReadQ(addrStr string, quantity uint16) ([]byte, error) {
	offset, area, err := s.parseAddress(addrStr)
	if err != nil || area != "Q" {
		return nil, fmt.Errorf("无效的Q区地址")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.client.ReadCoils(offset, quantity)
}

// WriteQ 写输出线圈，如 WriteQ("Q0.1", true)
func (s *Siemens) WriteQ(addrStr string, isOn bool) error {
	offset, area, err := s.parseAddress(addrStr)
	if err != nil || area != "Q" {
		return fmt.Errorf("无效的Q区地址")
	}

	var val uint16 = 0x0000
	if isOn {
		val = 0xFF00
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	_, err = s.client.WriteSingleCoil(offset, val)
	return err
}

// ReadI 读输入线圈，如 ReadI("I0.0", 8) 读 I0.0-I0.7
func (s *Siemens) ReadI(addrStr string, quantity uint16) ([]byte, error) {
	offset, area, err := s.parseAddress(addrStr)
	if err != nil || area != "I" {
		return nil, fmt.Errorf("无效的I区地址: %s", addrStr)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	// 使用 Modbus 功能码 02 (Read Discrete Inputs)
	return s.client.ReadDiscreteInputs(offset, quantity)
}
