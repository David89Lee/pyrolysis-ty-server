package record_service

import "github.com/David89Lee/pyrolysis_ty_server/model"

type RecordService struct {
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

func (t *RecordService) FindRecord(infoID uint) ([]*model.Record, error) {
	return model.FindRecord(infoID)
}
