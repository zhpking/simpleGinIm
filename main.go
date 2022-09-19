package main

import (
	"log"
	"simpleGinIm/helper"
	"simpleGinIm/router"
	"simpleGinIm/service"
	"sync"
)

func main() {
	// e := router.Router()
	// e.Run(":8080")

	wg := sync.WaitGroup{}
	wg.Add(1)

	// 启动api
	go StartApi()
	// 启动ws
	go StartWs()

	wg.Wait()
}

func StartApi() {
	port, err := helper.GetWsPort()
	if err != nil {
		log.Printf("[SYS CONFIG ERROR] %v\n", err)
	}

	// 启动退出登录队列
	go service.GetUserLoginOutByMQ()
	// 启动用户发送消息队列
	go service.UserSendMessageByMQ()

	e := router.Router()
	e.Run(":" + port)
}

func StartWs() {

	port, err := helper.GetApiPort()
	if err != nil {
		log.Printf("[SYS CONFIG ERROR] %v\n", err)
	}

	// 启动心跳检测
	go service.CheckWebSocketConn()

	e := router.Router()
	e.Run(":" + port)
}
