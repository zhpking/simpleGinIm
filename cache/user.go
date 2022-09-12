package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"simpleGinIm/define"
)

// 获取用户会话列表
func getUserRoomIdList(userId string, page, pageSize int64) []string {
	start := (page - 1) * pageSize
	end := page * pageSize
	return Redis.ZRange(context.Background(), define.REDIS_KEY_USER_ROOM_ID_LIST + userId, start, end).Val()
}

// 设置用户参与过的会话
func setUserRoomId(userId, roomId string, currentTime int64) {
	z := &redis.Z{
		Score:float64(currentTime),
		Member:roomId,
	}
	Redis.ZAdd(context.Background(), define.REDIS_KEY_USER_ROOM_ID_LIST + userId, z)
}
