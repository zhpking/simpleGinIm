package router

import (
	"github.com/gin-gonic/gin"
	"simpleGinIm/api/handler"
	"simpleGinIm/middleware"
)

// https://www.liwenzhou.com/posts/Go/struct2map/
func Router() *gin.Engine {
	r := gin.Default()
	// 获取登录token
	r.POST("/user/get_login_token", handler.GetLoginToken)

	// 校验登录中间件
	auth := r.Group("/u", middleware.AuthCheck())
	// 退出登录
	auth.POST("/user/login_out", handler.UserLoginOut)
	// 连接IM
	// auth.GET("/user/connect", service.Connect)
	// 推送消息
	// auth.POST("/user/push_message", handler.UserSendMessage)

	// 校验房间中间件
	roomAuth := r.Group("/rm", middleware.AuthCheck())
	// 创建群聊房间
	roomAuth.POST("/room/create_room", handler.CreateRoom)
	// 用户进入房间
	roomAuth.POST("/room/enter_room", handler.EnterRoom)
	// 单聊
	roomAuth.POST("/room/single_message", handler.SingleMessage)
	// 群聊
	roomAuth.POST("/room/room_message", handler.RoomMessage)
	// 邀请用户进入房间
	roomAuth.POST("/room/invite_user_enter_room", handler.InviteUserEnterRoom)
	// todo 同意进入房间
	// todo 拒绝进入房间
	// 退出房间
	roomAuth.POST("/room/exit_room", handler.ExitRoom)
	// 踢人
	roomAuth.POST("/room/kick_out_room", handler.KickOutRoom)
	// 获取会话列表
	roomAuth.GET("/room/get_room_list", handler.GetRoomList)
	// 获取某个会话聊天记录
	roomAuth.GET("/room/get_room_message_list", handler.GetRoomMessageList)
	// todo 添加好友
	// todo 删除好友
	// todo 获取好友列表

	return r
}
