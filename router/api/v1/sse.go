package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/David89Lee/pyrolysis_ty_server/common/sse"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary 采样数据实时推送（SSE）
// @Tags 采样数据
// @Produce text/event-stream
// @Success 200 {string} string "event: sample / data: {...}"
// @Router /api/v1/sample/stream [get]
func SampleStream(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // 关闭 nginx 缓冲，保证实时性

	events, unsubscribe := sse.DefaultHub.Subscribe()
	defer unsubscribe()

	zap.L().Info("SSE 客户端连接",
		zap.String("client", c.ClientIP()),
		zap.Int("clients", sse.DefaultHub.ClientCount()))

	// 心跳间隔：防止代理/浏览器因空闲断开连接
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// retry: 告诉浏览器断线后的重连间隔（毫秒）
	fmt.Fprint(c.Writer, "retry: 3000\n\n")

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			return false // 客户端断开
		case ev, ok := <-events:
			if !ok {
				return false
			}
			data, err := json.Marshal(ev)
			if err != nil {
				zap.L().Error("序列化采样数据失败", zap.Error(err))
				return true
			}
			fmt.Fprintf(w, "event: sample\ndata: %s\n\n", data)
			return true
		case <-ticker.C:
			fmt.Fprint(w, ": ping\n\n") // 注释行，仅用于保活
			return true
		}
	})

	zap.L().Info("SSE 客户端断开",
		zap.String("client", c.ClientIP()),
		zap.Int("clients", sse.DefaultHub.ClientCount()))
}
