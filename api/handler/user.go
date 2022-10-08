package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"simpleGinIm/define"
	"simpleGinIm/helper"
	"simpleGinIm/model"
	"strconv"
	"sync"
	"time"
)

func GetLoginToken(ctx *gin.Context) {
	userId := ctx.PostForm("user_id");
	currentTime := time.Now().Unix()

	// 获取该userId是否已经注册，如果没注册，入表
	user, err := model.GetUserByUserId(userId)
	if err == mongo.ErrNoDocuments {
		// 没有数据，相当于没注册，往用户表入一条数据
		user = &model.User{
			UserId:userId,
			LoginStatus:define.LOGIN_STATUS_OFFLINE,
			LastLoginTime:currentTime,
			LastLoginOutTime:0,
			UserStatus:define.USER_STATUS_OK,
			CreateAt:currentTime,
		}
		err = model.UserInsertOne(user)
		if err != nil {
			log.Printf("[DB ERROR]%v\n", err)
			helper.FailResponse(ctx, "系统错误")
			return
		}
	} else if err != nil {
		log.Printf("[DB ERROR]%v\n", err)
		helper.FailResponse(ctx, "系统错误")
		return
	} else if user.LoginStatus == 1 {
		helper.FailResponse(ctx, "用户已登录")
		return
	}

	token, err := helper.GenerateToken(userId)
	if err != nil {
		log.Printf("生成token失败,%v", err)
		helper.FailResponse(ctx, "获取token失败，请重试")
		return
	}

	// 获取连接信息
	wsIpList,_ := helper.GetWsAddress()
	wsPort, _ := helper.GetWsPort()
	tcpIpList, _ := helper.GetTcpAddress()
	tcpPort, _ := helper.GetTcpPort()

	data := map[string]string{
		"token":token,
		"tcpAddress":wsIpList[0],
		"tcpPort":wsPort,
		"websocketAddress":tcpIpList[0],
		"websocketPort":tcpPort,
	}

	helper.SucResponse(ctx, "suc", data)
}

func UserLoginOut(ctx *gin.Context) {
	userId := []string{ctx.PostForm("user_id")}
	currentTime := time.Now().Unix()
	// 更新用户信息
	err := model.UpdateUserLoginOutStatusByUserIdList(userId, define.LOGIN_STATUS_OFFLINE, currentTime)
	if err != nil {
		log.Printf("用户退出失败,err:%v", err)
		helper.FailResponse(ctx, "用户退出失败")
	} else {
		helper.SucResponse(ctx, "suc", make(map[string]string))

		// 获取所有websocket连接ip
		addressList, err := helper.GetWsAddress()
		if err != nil {
			log.Printf("[CONFIG ERROR] 获取wc地址错误，err:%v\n", err)
		}

		// 获取端口
		// port,err := helper.GetApiPort()
		port,err := helper.GetWsPort()
		if err != nil {
			log.Printf("[CONFIG ERROR] 获取api端口错误，err:%v\n", err)
		}

		// 群聊的话使用广播的方式发送退出信息
		toUserIdList, _ := json.Marshal(userId)
		messageType := define.MESSAGE_TYPE_DISCONNECT
		body := "to_user_id="+string(toUserIdList)+"&message_id=0&message_type="+strconv.Itoa(int(messageType))+"&message_data="
		wg := &sync.WaitGroup{}
		wg.Add(len(addressList))
		for _, address := range addressList {
			go func(address, body string, wg *sync.WaitGroup) {
				err = helper.SendPost("http://"+address+":"+port+"/user/push_message", body)
				if err != nil {
					log.Printf("[CONFIG ERROR] %v\n", err)
				}
				wg.Done()
			}(address, body, wg)
		}

		wg.Wait()
	}
}
