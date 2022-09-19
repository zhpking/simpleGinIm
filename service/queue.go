package service

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"log"
	"simpleGinIm/cache"
	"simpleGinIm/define"
	"time"
)

// 队列处理的内容，都放在这里

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
			err = UserLoginOut([]string{v})
			if err != nil {
				log.Println("[DB ERROR] %v\n")
			}
		}
	}
}

// 发送消息mq
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
				log.Printf("[JSON ERROR] queue.go, err:%v\n", err)
			}
			// todo 处理消息
			UserSendMessage(receiveMsg)
		}


		/*
		err = UserLoginOut(userId)
		if err != nil {
			log.Println("[DB ERROR] %v\n")
		}
		*/
	}
}

