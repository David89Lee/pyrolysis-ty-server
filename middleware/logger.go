package middleware

import (
	"os"

	"github.com/David89Lee/pyrolysis_ty_server/common"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger() {
	logWriter := &lumberjack.Logger{
		Filename:   common.LogSetting.Filename,
		MaxSize:    common.LogSetting.MaxSize,
		MaxBackups: common.LogSetting.MaxBackups,
		MaxAge:     common.LogSetting.MaxAge,
		Compress:   common.LogSetting.Compress,
	}

	// 配置 zap 使用 lumberjack
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// 控制台打印，不要可以注释掉
	consoleWriter := zapcore.AddSync(os.Stdout)

	fileWriter := zapcore.AddSync(logWriter)
	writeSyncer := zapcore.NewMultiWriteSyncer(consoleWriter, fileWriter)
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), writeSyncer, zap.InfoLevel)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	zap.ReplaceGlobals(logger)

	// 添加日志初始化信息
	zap.L().Info("logger initialized", zap.String("filename", common.LogSetting.Filename))
}
