package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

// Modbus 常量
const (
	FuncReadHoldingRegisters = 0x03 // 读保持寄存器
	FuncWriteSingleRegister  = 0x06 // 写单个寄存器
	FuncWriteMultipleRegs    = 0x10 // 写多个寄存器

	BroadcastAddr = 0xFF // 广播地址

	// 默认串口参数（厂家默认：9600, 8, N, 1）
	DefaultBaudRate = 9600
	DefaultDataBits = 8
	DefaultStopBits = 1

	// 设备响应超时（厂家要求 ≥500ms）
	DefaultTimeout = 1000 * time.Millisecond

	// 帧间间隔（MODBUS RTU 要求 ≥3.5 个字符时间）
	FrameDelay = 4 * time.Millisecond
)

// 常见错误
var (
	ErrInvalidCRC      = errors.New("modbus: CRC 校验失败")
	ErrInvalidResponse = errors.New("modbus: 无效的响应")
	ErrSlaveFailure    = errors.New("modbus: 从机返回异常码 (功能码最高位被置 1)")
)

// ModbusRTU 封装了一个基于串口的 MODBUS RTU 主站
type ModbusRTU struct {
	port     io.ReadWriteCloser
	timeout  time.Duration
	frameGap time.Duration
}

// NewModbusRTU 创建一个 MODBUS RTU 主站实例
// port 必须是已经配置好波特率/数据位/停止位/校验位的串口
func NewModbusRTU(port io.ReadWriteCloser, timeout time.Duration) *ModbusRTU {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return &ModbusRTU{
		port:     port,
		timeout:  timeout,
		frameGap: FrameDelay,
	}
}

// Port 返回底层串口（便于调用方再次配置或关闭）
func (m *ModbusRTU) Port() io.ReadWriteCloser { return m.port }

// SetTimeout 设置单次读写的超时
func (m *ModbusRTU) SetTimeout(t time.Duration) {
	if t > 0 {
		m.timeout = t
	}
}

// CRC16 计算 MODBUS 循环冗余校验（多项式 0xA001）
// 注意：MODBUS CRC 在链路上是低字节在前
func CRC16(data []byte) (lo, hi byte) {
	var crc uint16 = 0xFFFF
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&0x0001 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return byte(crc & 0xFF), byte((crc >> 8) & 0xFF)
}

// appendCRC 给 data 追加 MODBUS CRC（低字节在前）
func appendCRC(data []byte) []byte {
	lo, hi := CRC16(data)
	return append(data, lo, hi)
}

// verifyCRC 校验 data（不含 CRC）+ 末尾两字节 CRC 是否一致
func verifyCRC(frame []byte) bool {
	if len(frame) < 3 {
		return false
	}
	body := frame[:len(frame)-2]
	lo, hi := CRC16(body)
	return frame[len(frame)-2] == lo && frame[len(frame)-1] == hi
}

// sendRequest 发送一帧请求并清空接收缓冲区中的脏数据
func (m *ModbusRTU) sendRequest(frame []byte) error {
	// 帧间间隔，避免与上一帧粘连
	time.Sleep(m.frameGap)
	if _, err := m.port.Write(frame); err != nil {
		return fmt.Errorf("写入串口失败: %w", err)
	}
	return nil
}

// readResponse 读取从机响应，遇到 idleGap 长时间的静默即认为一帧结束
func (m *ModbusRTU) readResponse(minLen int, idleGap time.Duration) ([]byte, error) {
	buf := make([]byte, 256)
	total := 0
	lastRead := time.Now()

	for {
		// 超过整体超时
		if time.Since(lastRead) > m.timeout {
			break
		}
		// 静默 idleGap 视为一帧结束
		if total > 0 && time.Since(lastRead) > idleGap {
			break
		}
		n, err := m.port.Read(buf[total:])
		if n > 0 {
			total += n
			lastRead = time.Now()
			// 一旦满足最小长度并出现 idleGap 即可终止
			if total >= minLen {
				// 继续读到静默
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			// 串口读取超时（SetReadDeadline）通常以 timeout error 形式返回
			// 这里只在已经有数据时静默忽略，否则向上抛
			if total == 0 {
				return nil, fmt.Errorf("读取串口失败: %w", err)
			}
		}
		if total >= len(buf) {
			break
		}
	}

	if total == 0 {
		return nil, fmt.Errorf("从机无响应（超时 %v）", m.timeout)
	}
	return buf[:total], nil
}

// ReadHoldingRegisters 发送功能码 0x03 读保持寄存器，返回原始寄存器值（大端）
func (m *ModbusRTU) ReadHoldingRegisters(addr byte, startReg uint16, count uint16) ([]byte, error) {
	if count < 1 || count > 125 {
		return nil, fmt.Errorf("寄存器数量越界: %d", count)
	}
	req := []byte{addr, FuncReadHoldingRegisters}
	req = binary.BigEndian.AppendUint16(req, startReg)
	req = binary.BigEndian.AppendUint16(req, count)
	req = appendCRC(req)

	if err := m.sendRequest(req); err != nil {
		return nil, err
	}

	// 期望响应长度: addr(1) + func(1) + byteCount(1) + count*2 + crc(2)
	expectLen := 3 + int(count)*2 + 2
	resp, err := m.readResponse(expectLen, 5*time.Millisecond)
	if err != nil {
		return nil, err
	}
	if len(resp) < expectLen {
		return nil, fmt.Errorf("%w: 响应长度不足, 期望 %d 实际 %d", ErrInvalidResponse, expectLen, len(resp))
	}
	if !verifyCRC(resp[:expectLen]) {
		return nil, ErrInvalidCRC
	}
	// 异常响应：功能码最高位被置 1
	if resp[1]&0x80 != 0 {
		return nil, fmt.Errorf("%w: 异常码 0x%02X", ErrSlaveFailure, resp[2])
	}
	if resp[1] != FuncReadHoldingRegisters {
		return nil, fmt.Errorf("%w: 功能码不匹配 0x%02X", ErrInvalidResponse, resp[1])
	}
	byteCount := int(resp[2])
	if byteCount != int(count)*2 || len(resp) < 3+byteCount+2 {
		return nil, fmt.Errorf("%w: 字节数不匹配", ErrInvalidResponse)
	}
	data := make([]byte, byteCount)
	copy(data, resp[3:3+byteCount])
	return data, nil
}

// ReadU16 读取单个 16 位保持寄存器（无符号）
func (m *ModbusRTU) ReadU16(addr byte, reg uint16) (uint16, error) {
	data, err := m.ReadHoldingRegisters(addr, reg, 1)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(data[:2]), nil
}

// ReadI16 读取单个 16 位保持寄存器（有符号）
func (m *ModbusRTU) ReadI16(addr byte, reg uint16) (int16, error) {
	data, err := m.ReadHoldingRegisters(addr, reg, 1)
	if err != nil {
		return 0, err
	}
	return int16(binary.BigEndian.Uint16(data[:2])), nil
}

// WriteSingleRegister 发送功能码 0x06 写单个寄存器
func (m *ModbusRTU) WriteSingleRegister(addr byte, reg uint16, value uint16) error {
	req := []byte{addr, FuncWriteSingleRegister}
	req = binary.BigEndian.AppendUint16(req, reg)
	req = binary.BigEndian.AppendUint16(req, value)
	req = appendCRC(req)

	if err := m.sendRequest(req); err != nil {
		return err
	}

	// 广播地址不期待响应
	if addr == BroadcastAddr {
		return nil
	}

	// 写单寄存器响应长度固定 8 字节
	resp, err := m.readResponse(8, 5*time.Millisecond)
	if err != nil {
		return err
	}
	if len(resp) < 8 {
		return fmt.Errorf("%w: 响应长度不足, 期望 8 实际 %d", ErrInvalidResponse, len(resp))
	}
	if !verifyCRC(resp[:8]) {
		return ErrInvalidCRC
	}
	if resp[1]&0x80 != 0 {
		return fmt.Errorf("%w: 异常码 0x%02X", ErrSlaveFailure, resp[2])
	}
	if resp[1] != FuncWriteSingleRegister {
		return fmt.Errorf("%w: 功能码不匹配 0x%02X", ErrInvalidResponse, resp[1])
	}
	// 回显的寄存器地址与值应与请求一致
	gotReg := binary.BigEndian.Uint16(resp[2:4])
	gotVal := binary.BigEndian.Uint16(resp[4:6])
	if gotReg != reg || gotVal != value {
		return fmt.Errorf("%w: 回显内容不匹配 (reg=%d val=%d)", ErrInvalidResponse, gotReg, gotVal)
	}
	return nil
}

// WriteMultipleRegisters 发送功能码 0x10 写多个寄存器
func (m *ModbusRTU) WriteMultipleRegisters(addr byte, startReg uint16, values []uint16) error {
	if len(values) < 1 || len(values) > 123 {
		return fmt.Errorf("寄存器数量越界: %d", len(values))
	}
	req := []byte{addr, FuncWriteMultipleRegs}
	req = binary.BigEndian.AppendUint16(req, startReg)
	req = binary.BigEndian.AppendUint16(req, uint16(len(values)))
	req = append(req, byte(len(values)*2))
	for _, v := range values {
		req = binary.BigEndian.AppendUint16(req, v)
	}
	req = appendCRC(req)

	if err := m.sendRequest(req); err != nil {
		return err
	}

	if addr == BroadcastAddr {
		return nil
	}

	// 写多寄存器响应长度固定 8 字节
	resp, err := m.readResponse(8, 5*time.Millisecond)
	if err != nil {
		return err
	}
	if len(resp) < 8 {
		return fmt.Errorf("%w: 响应长度不足, 期望 8 实际 %d", ErrInvalidResponse, len(resp))
	}
	if !verifyCRC(resp[:8]) {
		return ErrInvalidCRC
	}
	if resp[1]&0x80 != 0 {
		return fmt.Errorf("%w: 异常码 0x%02X", ErrSlaveFailure, resp[2])
	}
	if resp[1] != FuncWriteMultipleRegs {
		return fmt.Errorf("%w: 功能码不匹配 0x%02X", ErrInvalidResponse, resp[1])
	}
	gotReg := binary.BigEndian.Uint16(resp[2:4])
	gotCnt := binary.BigEndian.Uint16(resp[4:6])
	if gotReg != startReg || int(gotCnt) != len(values) {
		return fmt.Errorf("%w: 回显内容不匹配 (reg=%d cnt=%d)", ErrInvalidResponse, gotReg, gotCnt)
	}
	return nil
}
