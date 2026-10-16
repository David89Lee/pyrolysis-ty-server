package response

import "github.com/gin-gonic/gin"

type Gin struct {
	C *gin.Context
}

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// ExcelFile 以文本（base64）方式承载 Excel 文件内容，
// 供第三方平台（如蓝卓 supOS）封装服务后无法直接下载二进制流时使用
type ExcelFile struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Body        string `json:"body"`
}

type Pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

type PageData struct {
	List       interface{} `json:"list"`
	Pagination `json:"pagination"`
}

func (g *Gin) Response(httpCode, errCode int, data interface{}) {
	g.C.JSON(httpCode, Response{
		Code: errCode,
		Msg:  GetMsg(errCode),
		Data: data,
	})
}

func (g *Gin) PageDataResponse(httpCode, errCode int, total int64, pageSize int, currentPage int, data interface{}) {
	g.C.JSON(httpCode, Response{
		Code: errCode,
		Msg:  GetMsg(errCode),
		Data: PageData{List: data, Pagination: Pagination{Page: currentPage, PageSize: pageSize, Total: total}},
	})
}
