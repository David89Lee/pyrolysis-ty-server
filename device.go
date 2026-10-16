package main

import (
	"encoding/binary"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// 寄存器地址表（参考 TY-RJ03-1 协议 V1.1）
const (
	RegDeviceAddr     uint16 = 0x0106 // 设备地址
	RegBaudRate       uint16 = 0x0001 // 波特率
	RegTemperature    uint16 = 0x0000 // 温度采样值
	RegHumidity       uint16 = 0x0003 // 湿度采样值
	RegPM1_0          uint16 = 0x0004 // PM1.0
	RegPM2_5          uint16 = 0x0005 // PM2.5
	RegPM10           uint16 = 0x0006 // PM10
	RegTVOCSample     uint16 = 0x0007 // TVOC 采样值
	RegTVOCBackground uint16 = 0x0008 // TVOC 本底值
	RegPyrolysis      uint16 = 0x0009 // 热解粒子采样值
	RegStatusCode     uint16 = 0x000A // 状态码

	// 阈值/复位寄存器（位于 0x0100 段）
	RegThrPM1_0       uint16 = 0x0100 // PM1.0 报警阈值
	RegThrPM2_5       uint16 = 0x0101 // PM2.5 报警阈值
	RegThrPM10        uint16 = 0x0102 // PM10 报警阈值
	RegThrTVOC        uint16 = 0x0103 // TVOC 报警阈值
	RegThrTemperature uint16 = 0x0104 // 温度报警阈值
	RegReset          uint16 = 0x0105 // 复位寄存器（写入 0x0001 触发复位）

	// 一次读取的采样寄存器起始与数量
	SampleStartReg = RegTemperature
	SampleRegCount = 9 // 0x0002 ~ 0x000A
)

// 状态码位定义
const (
	StatusAlarmMask       uint16 = 1 << 0 // Bit0: 0=未报警,   1=报警
	StatusSensorErrorMask uint16 = 1 << 1 // Bit1: 0=传感器正常, 1=传感器异常

)

// SampleData 保存一次完整采样的解析结果
type SampleData struct {
	Temperature    float64 // °C
	Humidity       float64 // %
	PM1_0          uint16  // μg/m³
	PM2_5          uint16  // μg/m³
	PM10           uint16  // μg/m³
	TVOCSample     uint16  // ppb
	TVOCBackground uint16  // ppb
	Pyrolysis      uint16  // ppb
	StatusCode     uint16  // 原始状态码
}

// SensorNormal 返回传感器是否正常
func (s SampleData) SensorNormal() bool {
	// 打印StatusSensorErrorMask
	//fmt.Printf("StatusSensorErrorMask: %v\n", StatusSensorErrorMask)
	return s.StatusCode&StatusSensorErrorMask == 0
}

// Alarming 返回是否处于报警状态
func (s SampleData) Alarming() bool {
	// 打印StatusAlarmMask
	//fmt.Printf("StatusAlarmMask: %v\n", StatusAlarmMask)
	return s.StatusCode&StatusAlarmMask != 0
}

// ReadSample 一次性读取 9 个采样寄存器并解析
// 协议示例（响应）:
//
//	01 03 12 03 02 01 70 00 0F 00 1A 00 12 00 00 00 87 00 00 00 00 [CRC]
//	数据部分: 温度=0x0302=770, 湿度=0x0170=368, PM1.0=0x000F=15,
//	          PM2.5=0x001A=26, PM10=0x0012=18, TVOC采样=0x0000=0,
//	          TVOC本底=0x0087=135, 热解粒子=0x0000=0, 状态码=0x0000=0
func (m *ModbusRTU) ReadSample(addr byte) (SampleData, error) {
	data, err := m.ReadHoldingRegisters(addr, SampleStartReg, SampleRegCount)
	if err != nil {
		return SampleData{}, err
	}
	if len(data) < SampleRegCount*2 {
		return SampleData{}, fmt.Errorf("采样数据长度不足: %d", len(data))
	}
	regs := make([]uint16, SampleRegCount)
	for i := 0; i < SampleRegCount; i++ {
		regs[i] = binary.BigEndian.Uint16(data[i*2 : i*2+2])
	}
	// 温度: (寄存器值 - 500) / 10  °C
	temp := float64(int16(regs[0])) - 500
	temp /= 10
	// 湿度: 寄存器值 / 10  %
	hum := float64(int16(regs[1])) / 10
	return SampleData{
		Temperature:    temp,
		Humidity:       hum,
		PM1_0:          regs[2],
		PM2_5:          regs[3],
		PM10:           regs[4],
		TVOCSample:     regs[5],
		TVOCBackground: regs[6],
		Pyrolysis:      regs[7],
		StatusCode:     regs[8],
	}, nil
}

// Format 返回一行带单位的人类可读字符串
func (s SampleData) Format() string {
	sensor := "正常"
	if !s.SensorNormal() {
		sensor = "异常"
	}
	alarm := "未报警"
	if s.Alarming() {
		alarm = "报警"
	}
	return fmt.Sprintf(
		"温度=%.1f°C  湿度=%.1f%%  PM1.0=%dμg/m³  PM2.5=%dμg/m³  PM10=%dμg/m³  "+
			"TVOC采样=%dppb  TVOC本底=%dppb  热解粒子=%dppb  状态码=0x%04X(传感器%s,%s)",
		s.Temperature, s.Humidity, s.PM1_0, s.PM2_5, s.PM10,
		s.TVOCSample, s.TVOCBackground, s.Pyrolysis, s.StatusCode, sensor, alarm,
	)
}

// ===== 写入功能（可在外部调用） =====

// SetDeviceAddress 广播修改从机地址（使用广播地址 0xFF）
// 注意: 本设备广播修改地址使用自定义帧格式，并非标准 MODBUS 写单寄存器。
// 协议示例（PDF 第 3 节）: FF 03 01 06 00 01 70 29  （将地址改为 1）
//   - FF: 广播地址
//   - 03: 自定义功能码（厂家用作"广播修改设备地址"）
//   - 01 06: 固定前缀（"0106 为广播地址命令"）
//   - 00 [addr]: 新地址（16 位）
//   - 末 2 字节: MODBUS CRC-16
func (m *ModbusRTU) SetDeviceAddress(newAddr byte) error {
	if newAddr == 0 || newAddr == BroadcastAddr {
		return fmt.Errorf("非法的设备地址: 0x%02X", newAddr)
	}
	frame := []byte{BroadcastAddr, 0x03, 0x01, 0x06, 0x00, newAddr}
	frame = appendCRC(frame)
	if err := m.sendRequest(frame); err != nil {
		return err
	}
	// 广播不期待响应
	return nil
}

// SetBaudRate 修改波特率寄存器（具体编码请以厂家手册为准，常见为 0=1200,1=2400,2=4800,3=9600）
// 注意：修改波特率后需同步调整本机串口配置
func (m *ModbusRTU) SetBaudRate(addr byte, code uint16) error {
	return m.WriteSingleRegister(addr, RegBaudRate, code)
}

// SetThresholds 批量写入 PM1.0/PM2.5/PM10/TVOC/温度 报警阈值
// 协议示例（写 5 个寄存器）:
//
//	01 10 01 00 00 05 0A 00 64 00 64 00 64 00 64 00 1E [CRC]
func (m *ModbusRTU) SetThresholds(addr byte, pm1_0, pm2_5, pm10, tvoc, temperature uint16) error {
	return m.WriteMultipleRegisters(addr, RegThrPM1_0,
		[]uint16{pm1_0, pm2_5, pm10, tvoc, temperature})
}

// ResetDevice 写入 0x0001 到复位寄存器，触发设备复位
// 协议示例: 01 10 01 05 00 01 02 00 01 [CRC]
func (m *ModbusRTU) ResetDevice(addr byte) error {
	return m.WriteMultipleRegisters(addr, RegReset, []uint16{0x0001})
}

// ReadDeviceAddress 读取设备地址寄存器
func (m *ModbusRTU) ReadDeviceAddress(addr byte) (byte, error) {
	v, err := m.ReadU16(addr, RegDeviceAddr)
	return byte(v), err
}

// ReadBaudRate 读取波特率寄存器
func (m *ModbusRTU) ReadBaudRate(addr byte) (uint16, error) {
	return m.ReadU16(addr, RegBaudRate)
}

// ReadThresholds 读取 5 个报警阈值（PM1.0/PM2.5/PM10/TVOC/温度）
func (m *ModbusRTU) ReadThresholds(addr byte) (pm1_0, pm2_5, pm10, tvoc, temperature uint16, err error) {
	data, err := m.ReadHoldingRegisters(addr, RegThrPM1_0, 5)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	pm1_0 = binary.BigEndian.Uint16(data[0:2])
	pm2_5 = binary.BigEndian.Uint16(data[2:4])
	pm10 = binary.BigEndian.Uint16(data[4:6])
	tvoc = binary.BigEndian.Uint16(data[6:8])
	temperature = binary.BigEndian.Uint16(data[8:10])
	return
}

// PollLoop 周期性采样并打印，interval 为采样周期
func (m *ModbusRTU) PollLoop(addr byte, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		sample, err := m.ReadSample(addr)
		ts := time.Now().Format("2006-01-02 15:04:05")
		zap.L().Info("开始读取", zap.String("time", ts))
		if err != nil {
			zap.L().Error("读取失败", zap.Error(err))
		} else {
			zap.L().Info("读取成功", zap.String("sample", sample.Format()))
		}
		<-ticker.C
	}
}
