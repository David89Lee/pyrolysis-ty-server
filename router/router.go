package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	_ "github.com/David89Lee/pyrolysis_ty_server/docs"
	v1 "github.com/David89Lee/pyrolysis_ty_server/router/api/v1"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	r.Use(gin.Recovery())

	// 使用zap记录日志
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		zap.L().Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
		)
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	apiv1 := r.Group("/api/v1")

	apiv1.POST("/info", v1.CreateInfo)
	apiv1.PUT("/info", v1.UpdateInfo)
	apiv1.DELETE("/info/:id", v1.DeleteInfo)
	apiv1.GET("/info", v1.FindInfo)

	return r
}
