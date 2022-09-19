package router

import (
	"github.com/gin-gonic/gin"
	"simpleGinIm/middleware"
	"simpleGinIm/service"
)

// https://www.liwenzhou.com/posts/Go/struct2map/
func Router() *gin.Engine {
	r := gin.Default()
	// 获取登录token
	r.POST("/user/get_login_token", service.GetLoginToken)

	// 校验登录中间件
	auth := r.Group("/u", middleware.AuthCheck())
	// 退出登录
	auth.POST("/user/login_out", service.LoginOut)
	// 连接IM
	auth.GET("/user/connect", service.Connect)
	// 推送消息
	r.POST("/user/push_message", service.PushSingleMessage)

	// 校验房间中间件
	roomAuth := r.Group("/rm", middleware.AuthCheck())
	// 创建群聊房间
	roomAuth.POST("/room/create_room", service.CreateRoom)
	// 用户进入房间
	roomAuth.POST("/room/enter_room", service.EnterRoom)
	// 邀请用户进入房间
	roomAuth.POST("/room/invite_user_enter_room", service.InviteUserEnterRoom)
	// todo 同意进入房间
	// todo 拒绝进入房间
	// 退出房间
	roomAuth.POST("/room/exit_room", service.ExitRoom)
	// 踢人
	roomAuth.POST("/room/kick_out_room", service.KickOutRoom)
	// 获取会话列表
	roomAuth.GET("/room/get_room_list", service.GetRoomList)
	// 获取某个会话聊天记录
	roomAuth.GET("/room/get_room_message_list", service.GetRoomMessageList)
	// todo 添加好友
	// todo 删除好友
	// todo 获取好友列表

	return r
}
