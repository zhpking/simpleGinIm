package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"simpleGinIm/define"
)

// 获取用户会话列表
func GetUserRoomIdList(userId string, page, pageSize int64) []string {
	start := (page - 1) * pageSize
	end := page * pageSize
	return Redis.ZRange(context.Background(), define.REDIS_SET_USER_ROOM_ID_LIST + userId, start, end).Val()
}

// 设置用户参与过的会话
func SetUserRoomId(userId, roomId string, currentTime int64) {
	z := &redis.Z{
		Score:float64(currentTime),
		Member:roomId,
	}
	Redis.ZAdd(context.Background(), define.REDIS_SET_USER_ROOM_ID_LIST + userId, z)
}

// 获取用户长连接路由
func GetConnection2User(userId string) {
	Redis.Get(context.Background(), define.REDIS_STRING_USER_WEBSOCKET_CONNECT + userId)
}

// 设置用户长连接路由
func SetConnection2User(userId, ipAddress string) {
	Redis.Set(context.Background(), define.REDIS_STRING_USER_WEBSOCKET_CONNECT + userId, ipAddress, 0)
}

// 设置退出登录用户消息队列
func SetUserLoginOutList(userId string) {
	Redis.LPush(context.Background(), define.REDIS_LIST_USER_LOGIN_OUT, userId)
}

// 设置用户发送消息消息队列
func SetUserSendMessageList(data string) {
	Redis.LPush(context.Background(), define.REDIS_LIST_USER_SEND_MESSAGE, data)
}