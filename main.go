package main

import (
	"fmt"

	"github.com/David89Lee/pyrolysis_ty_server/common"
	"github.com/David89Lee/pyrolysis_ty_server/middleware"
	"github.com/David89Lee/pyrolysis_ty_server/model"
	"github.com/David89Lee/pyrolysis_ty_server/router"
	"go.uber.org/zap"

	"flag"
	"os"
	"time"

	"go.bug.st/serial"
)

func init() {
	common.SetUpSetting("conf")
	middleware.InitLogger()
	model.Setup()
}

// @title 台谊热解粒子测试
// @version 1.0
// @description 台谊热解粒子测试
// @contact.name LG
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host 127.0.0.1:8080

func main() {

	defer func(l *zap.Logger) {
		err := l.Sync()
		if err != nil {
			zap.L().Error("zap日志同步失败", zap.Error(err))
		}
	}(zap.L())
	// http服务
	r := router.InitRouter()
	// start server in a goroutine
	go func() {
		if err := r.Run(common.ServerSetting.HttpPort); err != nil {
			zap.L().Error("启动服务失败", zap.Error(err))
		}
	}()

	conn()
}

// 默认参数（可通过命令行覆盖）
const (
	defaultSerialPort = "COM4"          // Windows 下默认串口
	defaultSlaveAddr  = byte(0x02)      // 厂家默认从机地址
	defaultPollCycle  = 5 * time.Second // 采样周期
)

func conn() {
	// ===== 命令行参数 =====
	port := flag.String("port", defaultSerialPort, "串口设备路径 (Windows: COM3, Linux: /dev/ttyUSB0)")
	baud := flag.Int("baud", DefaultBaudRate, "波特率")
	addr := flag.Int("addr", int(defaultSlaveAddr), "从机地址 (1~247)")
	interval := flag.Duration("interval", defaultPollCycle, "采样周期")
	timeout := flag.Duration("timeout", DefaultTimeout, "单次通信超时")
	flag.Parse()

	if *addr < 1 || *addr > 247 {
		zap.L().Error("从机地址越界（1~247）", zap.Int("addr", *addr))
		os.Exit(2)
	}

	// ===== 打开并配置串口（厂家默认: 9600, 8, N, 1） =====
	mode := &serial.Mode{
		BaudRate: *baud,
		DataBits: DefaultDataBits,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
	com, err := serial.Open(*port, mode)
	if err != nil {
		zap.L().Error("打开串口失败", zap.String("port", *port), zap.Error(err))
		os.Exit(1)
	}
	defer com.Close()
	// 设置单次读取超时，避免 Read 长时间阻塞
	_ = com.SetReadTimeout(*timeout)

	mb := NewModbusRTU(com, *timeout)

	// ===== 启动周期采样 =====
	slave := byte(*addr)
	zap.L().Info("开始采集",
		zap.String("port", *port),
		zap.Int("baud", *baud),
		zap.Any("slave", slave),
		zap.Duration("interval", *interval),
		zap.Duration("timeout", *timeout))

	// 读取设备地址
	deviceAddr, err := mb.ReadDeviceAddress(0xFF)
	if err != nil {
		zap.L().Error("读取设备地址失败", zap.Error(err))
		os.Exit(1)
	}
	fmt.Printf("设备地址: %d\n", deviceAddr)

	// 立即先读取一次，确认链路正常
	if sample, err := mb.ReadSample(slave); err != nil {
		zap.L().Error("首次读取失败", zap.Error(err))
		zap.L().Error("请检查: 串口路径/权限、波特率、从机地址、RS485 接线与换向方向")
		os.Exit(1)
	} else {
		fmt.Printf("[首次] %s\n", sample.Format())
	}

	// 读取一下报警阈值
	if pm1_0, pm2_5, pm10, tvoc, temperature, err := mb.ReadThresholds(slave); err != nil {
		zap.L().Error("读取报警阈值失败", zap.Error(err))
	} else {
		zap.L().Info("读取报警阈值成功",
			zap.Uint16("PM1.0", pm1_0),
			zap.Uint16("PM2.5", pm2_5),
			zap.Uint16("PM10", pm10),
			zap.Uint16("TVOC", tvoc),
			zap.Uint16("Temperature", temperature))
	}

	mb.PollLoop(slave, *interval)
}
