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
