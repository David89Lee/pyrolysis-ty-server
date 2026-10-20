package model

import (
	"time"

	"gorm.io/gorm"
)

type Info struct {
	gorm.Model
	Description string `Gorm:"type:varchar(255); comment:'描述'"`
}
type Record struct {
	ID             uint `gorm:"primary_key"`
	CreatedAt      time.Time
	InfoID         uint
	Temperature    float32 `Gorm:"type:float; comment '温度 °C'"`
	Humidity       float32 `Gorm:"type:float; comment '湿度 %'"`
	Pm1            int16   `Gorm:"type:smallint; comment:'pm1.0 μg/m³'"`
	Pm2dot5        int16   `Gorm:"type:smallint; comment:'pm2.5 μg/m³'"`
	Pm10           int16   `Gorm:"type:smallint; comment:'pm10 μg/m³'"`
	TvocSampling   int16   `Gorm:"type:smallint; comment:'tvoc采样值 ppb'"`
	TvocBackground int16   `Gorm:"type:smallint; comment:'tvoc本底值 ppb'"`
	Pyrolysis      int16   `Gorm:"type:smallint; comment:'热解粒子值 ppb'"`
	SensorState    bool    `Gorm:" comment:'传感器状态'"`
	Alarm          bool    `Gorm:" comment:'是否报警'"`
}

func CreateInfo(info *Info) (*Info, error) {
	result := DB.Create(info)
	return info, result.Error
}

func UpdateInfo(info *Info) (*Info, error) {
	result := DB.Save(info)
	return info, result.Error
}

func DeleteInfo(id uint) error {
	return DB.Delete(&Info{}, id).Error
}

func FindInfo(pageNum, pageSize int) ([]*Info, int64, error) {
	db := DB.Model(&Info{})
	var total int64
	db.Count(&total)
	var info []*Info
	err := db.Order("id desc").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&info).Error
	return info, total, err
}

func CreateRecord(record *Record) (*Record, error) {
	result := DB.Create(record)
	return record, result.Error
}

func FindRecord(pageNum, pageSize int, infoID uint) ([]*Record, int64, error) {
	db := DB.Model(&Record{})
	db = db.Where("info_id = ?", infoID)
	var total int64
	db.Count(&total)
	var record []*Record
	err := db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&record).Error
	return record, total, err
}
