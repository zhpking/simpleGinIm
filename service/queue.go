package service

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
	"simpleGinIm/cache"
	"simpleGinIm/define"
	"simpleGinIm/model"
	"time"
)

// 队列处理的内容，都放在这里

// 登录mq
func GetUserLoginByMQ() {
	for {
		// 设置一个5秒的超时时间
		userId, err := cache.Redis.BRPop(context.Background(), 5 * time.Second,define.REDIS_LIST_USER_LOGIN).Result()
		if err == redis.Nil{
			// 查询不到数据
			time.Sleep(1 * time.Second)
			continue
		}
		if err != nil {
			// 查询出错
			time.Sleep(1 * time.Second)
			continue
		}

		currentTime := time.Now().Unix()
		userIdList := make([]string, 0)
		for _, v := range userId {
			// 不知道为啥data里面会带着key值，这里过滤下
			if v == define.REDIS_LIST_USER_LOGIN {
				continue
			}

			userIdList = append(userIdList, v)
		}

		log.Println(userIdList, "登录mq")
		if len(userIdList) > 0 {
			// 修改用户登录时间等信息
			err = model.UpdateUserLoginStatusByUserIdList(userIdList, define.LOGIN_STATUS_ONLINE, currentTime)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

// 退出登录mq
func GetUserLoginOutByMQ() {
	for {
		// 设置一个5秒的超时时间
		userId, err := cache.Redis.BRPop(context.Background(), 5 * time.Second,define.REDIS_LIST_USER_LOGIN_OUT).Result()
		if err == redis.Nil{
			// 查询不到数据
			time.Sleep(1 * time.Second)
			continue
		}
		if err != nil {
			// 查询出错
			time.Sleep(1 * time.Second)
			continue
		}

		for _, v := range userId {
			// 不知道为啥data里面会带着key值，这里过滤下
			if v == define.REDIS_LIST_USER_SEND_MESSAGE {
				continue
			}

			if err != nil {
				log.Println("[DB ERROR] %v\n")
			}
		}
	}
}

// 发送消息mq
/*
func UserSendMessageByMQ() {
	for {
		// 设置一个5秒的超时时间
		data, err := cache.Redis.BRPop(context.Background(), 5 * time.Second, define.REDIS_LIST_USER_SEND_MESSAGE).Result()
		if err == redis.Nil{
			// 查询不到数据
			time.Sleep(1 * time.Second)
			continue
		}
		if err != nil {
			// 查询出错
			time.Sleep(1 * time.Second)
			continue
		}


		for _, v := range data {
			// 不知道为啥data里面会带着key值，这里过滤下
			if v == define.REDIS_LIST_USER_SEND_MESSAGE {
				continue
			}
			receiveMsg := &define.ReceiveMessage{}
			err = json.Unmarshal([]byte(v), receiveMsg)
			if err != nil {
				log.Printf("[JSON ERROR] send message queue, err:%v\n", err)
			}
			// todo 处理消息
			// UserSendMessage(receiveMsg)
		}
	}
}
*/
