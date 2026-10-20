package v1

import (
	"net/http"
	"strconv"

	"github.com/David89Lee/pyrolysis_ty_server/model"
	"github.com/David89Lee/pyrolysis_ty_server/router/response"
	"github.com/David89Lee/pyrolysis_ty_server/service/record_service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary 信息
// @Tags 全信息查询
// @Accept json
// @Produce  json
// @Success 200 {object} response.Response{data=[]model.Info} "成功"
// @Failure 500 {object} response.Response
// @Router /api/v1/info [get]
func FindInfo(c *gin.Context) {
	appG := response.Gin{C: c}
	page := c.Query("page")
	pageSize := c.Query("pageSize")

	pageNum, _ := strconv.Atoi(page)
	pageSizeNum, _ := strconv.Atoi(pageSize)

	recordService := record_service.RecordService{}
	infos, total, err := recordService.FindInfo(pageNum, pageSizeNum)
	if err != nil {
		zap.L().Error("fail to find info", zap.Error(err))
		appG.Response(http.StatusOK, response.ErrorFindInfo, err)
		return
	}
	appG.PageDataResponse(http.StatusOK, response.SUCCESS, total, pageSizeNum, pageNum, infos)
}

// @Summary 信息
// @Tags 创建信息
// @Accept json
// @Produce  json
// @Param info body model.Info true "信息"
// @Success 200 {object} response.Response{data=model.Info}
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response
// @Router /api/v1/info [post]
func CreateInfo(c *gin.Context) {
	appG := response.Gin{C: c}

	info := &model.Info{}
	err := c.BindJSON(info)
	if err != nil {
		zap.L().Error("BindJSON error", zap.Error(err))
		appG.Response(http.StatusOK, response.InvalidParams, err)
		return
	}
	recordService := record_service.RecordService{}

	info, err = recordService.CreateInfo(info)
	if err != nil {
		zap.L().Error("fail to create info", zap.Error(err))
		appG.Response(http.StatusOK, response.ErrorCreateInfo, err)
		return
	}
	appG.Response(http.StatusOK, response.SUCCESS, info)
}

// @Summary 信息
// @Tags 更新信息
// @Accept json
// @Produce  json
// @Param info body model.Info true "信息"
// @Success 200 {object} response.Response{data=model.Info}
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response
// @Router /api/v1/info [put]
func UpdateInfo(c *gin.Context) {
	appG := response.Gin{C: c}

	info := &model.Info{}
	err := c.BindJSON(info)
	if err != nil {
		zap.L().Error("BindJSON error", zap.Error(err))
		appG.Response(http.StatusOK, response.InvalidParams, err)
		return
	}
	recordService := record_service.RecordService{}

	info, err = recordService.UpdateInfo(info)
	if err != nil {
		zap.L().Error("fail to update info", zap.Error(err))
		appG.Response(http.StatusOK, response.ErrorUpdateInfo, err)
		return
	}
	appG.Response(http.StatusOK, response.SUCCESS, info)
}

// @Summary 信息
// @Tags 删除信息
// @Accept json
// @Produce  json
// @Param id path int true "信息id"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response
// @Router /api/v1/info/{id} [delete]
func DeleteInfo(c *gin.Context) {
	appG := response.Gin{C: c}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		zap.L().Error("parse id error")
		appG.Response(response.ServerError, response.InvalidParams, err.Error())
		return
	}

	recordService := record_service.RecordService{}
	err = recordService.DeleteInfo(uint(id))
	if err != nil {
		zap.L().Error("fail to delete info", zap.Error(err))
		appG.Response(response.ServerError, response.ErrorDeleteInfo, err.Error())
		return
	}

	appG.Response(http.StatusOK, response.SUCCESS, nil)
}

// @Summary 记录
// @Tags 开启/停止记录数据
// @Accept json
// @Produce  json
// @Param id path int true "信息id；>0 开始记录（作为 model.Record.InfoID 落库），=0 停止记录"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response
// @Router /api/v1/info/{id}/record [post]
func Record(c *gin.Context) {
	appG := response.Gin{C: c}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		zap.L().Error("parse id error", zap.Error(err))
		appG.Response(http.StatusOK, response.InvalidParams, err.Error())
		return
	}
	if id < 0 {
		zap.L().Error("非法的记录id", zap.Int("id", id))
		appG.Response(http.StatusOK, response.InvalidParams, "id不能为负数")
		return
	}

	// id > 0：开始记录，之后 ReadSample 读到的每条采样数据
	// 都会以 id 作为 model.Record.InfoID 调用 CreateRecord 写入数据库
	if id > 0 {
		record_service.SetRecordInfoID(uint(id))
		zap.L().Info("开始记录采样数据", zap.Int("infoID", id))
		appG.Response(http.StatusOK, response.SUCCESS, gin.H{"info_id": id, "recording": true})
		return
	}

	// id = 0：停止记录
	record_service.SetRecordInfoID(0)
	zap.L().Info("停止记录采样数据")
	appG.Response(http.StatusOK, response.SUCCESS, gin.H{"info_id": 0, "recording": false})
}

// @Summary 记录
// @Tags 获取记录数据
// @Accept json
// @Produce  json
// @Param id path int true "model.Record.InfoID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response
// @Router /api/v1/info/{id}/record [get]
func FindRecord(c *gin.Context) {
	appG := response.Gin{C: c}

	page := c.Query("page")
	pageSize := c.Query("pageSize")

	pageNum, _ := strconv.Atoi(page)
	pageSizeNum, _ := strconv.Atoi(pageSize)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		zap.L().Error("parse id error", zap.Error(err))
		appG.Response(http.StatusOK, response.InvalidParams, err.Error())
		return
	}
	recordService := record_service.RecordService{}
	records, total, err := recordService.FindRecord(pageNum, pageSizeNum, uint(id))
	if err != nil {
		zap.L().Error("fail to find record", zap.Error(err))
		appG.Response(http.StatusOK, response.ErrorFindRecord, err)
		return
	}
	appG.PageDataResponse(http.StatusOK, response.SUCCESS, total, 0, 0, records)
}
