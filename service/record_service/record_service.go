package record_service

import (
	"sync"

	"github.com/David89Lee/pyrolysis_ty_server/model"
)

type RecordService struct {
}

// ===== 采样数据落库开关 =====
//
// recordInfoID 保存当前正在记录的 InfoID（即 model.Record.InfoID）：
//   - > 0：开启记录，ReadSample 读到的每条采样数据都会写入数据库
//   - = 0：停止记录
//
// 采集协程与 HTTP 协程都会访问该变量，故用读写锁保护。
var (
	recordMu     sync.RWMutex
	recordInfoID uint
)

// SetRecordInfoID 控制采样数据是否落库：id > 0 开始记录，id = 0 停止记录
func SetRecordInfoID(id uint) {
	recordMu.Lock()
	recordInfoID = id
	recordMu.Unlock()
}

// RecordInfoID 返回当前落库使用的 InfoID，0 表示未开启记录
func RecordInfoID() uint {
	recordMu.RLock()
	defer recordMu.RUnlock()
	return recordInfoID
}

// RecordSwitchOn 返回当前是否处于记录中（InfoID > 0）
func RecordSwitchOn() bool {
	return RecordInfoID() > 0
}

func (t *RecordService) CreateInfo(info *model.Info) (*model.Info, error) {
	return model.CreateInfo(info)
}

func (t *RecordService) UpdateInfo(info *model.Info) (*model.Info, error) {
	return model.UpdateInfo(info)
}

func (t *RecordService) DeleteInfo(id uint) error {
	return model.DeleteInfo(id)
}
func (t *RecordService) FindInfo(pageNum, pageSize int) ([]*model.Info, int64, error) {
	return model.FindInfo(pageNum, pageSize)
}

func (t *RecordService) CreateRecord(record *model.Record) (*model.Record, error) {
	return model.CreateRecord(record)
}

func (t *RecordService) FindRecord(pageNum, pageSize int, infoID uint) ([]*model.Record, int64, error) {
	return model.FindRecord(pageNum, pageSize, infoID)
}
