package main

import (
	"log"
	"simpleGinIm/connect/handler"
	"simpleGinIm/example"
	"simpleGinIm/helper"
	"simpleGinIm/api/router"
	router2 "simpleGinIm/connect/router"
	"simpleGinIm/service"
	"sync"
	"time"
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

	go StartTcp()
	time.Sleep(3 * time.Second)
	example.TcpClient()

	wg.Wait()
}

func StartApi() {
	port, err := helper.GetApiPort()
	if err != nil {
		log.Printf("[SYS CONFIG ERROR] %v\n", err)
	}

	// 启动用户登录队列
	go service.GetUserLoginByMQ()
	// 启动用户退出登录队列
	go service.GetUserLoginOutByMQ()
	// 启动用户发送处理消息队列
	// go service.UserSendMessageByMQ()
	// 服务注册
	// go service.RegisterService()

	e := router.Router()
	e.Run(":" + port)
}

func StartWs() {

	port, err := helper.GetWsPort()
	if err != nil {
		log.Printf("[SYS CONFIG ERROR] %v\n", err)
	}
	// port = "12345"


	// 启动心跳检测
	go handler.CheckWebSocketConn()

	e := router2.Router()
	e.Run(":" + port)
}

func StartTcp() {
	// 心跳检测
	// go handler.CheckTcpConn()
	// 检测tcp临时连接
	// go handler.CheckConnectTsConn()
	handler.TcpConnect()
}
