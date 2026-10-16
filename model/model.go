package model

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/David89Lee/pyrolysis_ty_server/common"
)

var DB *gorm.DB

type GormLogger struct{}

func (*GormLogger) Print(v ...interface{}) {
	switch v[0] {
	case "sql":
		zap.L().With(
			zap.String("module", "gorm"),
			zap.String("type", "sql"),
			zap.Int64("rows", v[5].(int64)),
			zap.String("src_ref", v[1].(string)),
			zap.Any("values", v[4]),
		).Info(v[3].(string))
	case "log":
		zap.L().With(
			zap.String("module", "gorm"),
			zap.String("type", "log"),
		).Info("")
	}
}

func (*GormLogger) Printf(s string, i ...interface{}) {
	zap.L().Sugar().Info(s, i)
}

func Setup() {
	var err error
	zap.L().With(
		zap.String("Type", common.DatabaseSetting.Type),
		zap.String("User", common.DatabaseSetting.User),
		zap.String("Password", common.DatabaseSetting.Password),
		zap.String("Host", common.DatabaseSetting.Host),
		zap.String("Name", common.DatabaseSetting.Name),
	).Debug("")

	gormLogger := &GormLogger{}

	// gorm初始化
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		common.DatabaseSetting.User,
		common.DatabaseSetting.Password,
		common.DatabaseSetting.Host,
		common.DatabaseSetting.Name)
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.New(gormLogger, logger.Config{
			//SlowThreshold:             2 * time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		}),
	})

	if err != nil {
		zap.L().Fatal("Gorm初始化失败", zap.Error(err))
	}

	err = DB.AutoMigrate()
	if err != nil {
		zap.L().Error("Gorm自动迁移失败", zap.Error(err))
	}
}
