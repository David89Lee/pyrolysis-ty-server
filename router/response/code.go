package response

const (
	SUCCESS                    = 200
	ERROR                      = 500
	InvalidParams              = 400
	ServerError                = 600
	ErrorAuthCheckTokenFail    = 601
	ErrorAuthCheckTokenTimeout = 602
	ErrorGenerateToken         = 603
	ErrorConnectWebSocket      = 610
	ErrorReadWebSocket         = 611

	ErrorCreateInfo = 10001
	ErrorUpdateInfo = 10002
	ErrorDeleteInfo = 10003
	ErrorFindInfo   = 10004
)
