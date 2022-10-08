package router

import (
	"github.com/gin-gonic/gin"
	"simpleGinIm/connect/handler"
	"simpleGinIm/middleware"
)

// https://www.liwenzhou.com/posts/Go/struct2map/
func Router() *gin.Engine {
	r := gin.Default()
	// 校验登录中间件
	auth := r.Group("/u", middleware.AuthCheck())
	// 连接IM (websocket)
	auth.GET("/user/connect", handler.Connect)
	// 推送消息
	r.POST("/user/push_message", handler.PushSingleMessage)

	return r
}
