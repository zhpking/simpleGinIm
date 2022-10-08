package handler

import (
	"context"
	"encoding/json"
	"log"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"simpleGinIm/cache"
	"simpleGinIm/define"
	"simpleGinIm/helper"
	"simpleGinIm/model"
	"strconv"
	"sync"
	"time"
)

func SingleMessage(ctx *gin.Context) {
	message := ctx.PostForm("message")
	toUserId := ctx.PostForm("to_user_id")
	userToken := ctx.MustGet("user_token").(*helper.UserToken)
	currentTime := time.Now().Unix()

	// 新建一条消息
	msg := &model.Message{
		MessageId:helper.GetUuid(),
		MessageData:message,
		MessageType:define.MESSAGE_TYPE_TEXT,
		MessageStatus:define.MESSAGE_STATUS_OK,
		CreateAt:currentTime,
	}
	err := model.MessageInsertOne(msg)
	if err != nil {
		log.Printf("[DB ERROR] %v\n", err)
		helper.FailResponse(ctx, "系统错误")
		return
	}

	// 两人是否拥有房间
	isHasRoom := true

	// 获取单聊房间
	sendUrList, err := model.GetUserRoomListByUserId(userToken.UserId, define.ROOM_TYPE_SINGLE)
	sendRoomIdList := make([]string, 0)
	if err == mongo.ErrNoDocuments {
		isHasRoom = false
	} else {
		for _, v := range sendUrList {
			sendRoomIdList = append(sendRoomIdList, v.RoomId)
		}
	}

	var toUrList []*model.UserRoom
	toRoomIdList := make([]string, 0)
	if isHasRoom {
		toUrList, err = model.GetUserRoomListByUserId(toUserId, define.ROOM_TYPE_SINGLE)
		if err == mongo.ErrNoDocuments {
			isHasRoom = false
		} else {
			for _, v := range toUrList {
				toRoomIdList = append(toRoomIdList, v.RoomId)
			}
		}
	}

	singleRoomId := ""
	if isHasRoom {
		// 查找房间并集
		intersectList := helper.IntersectArray(sendRoomIdList, toRoomIdList)
		if len(intersectList) > 0 {
			singleRoomId = intersectList[0]
		}
	}

	if singleRoomId == "" {
		// 没有私聊房间，创建一个私聊房间
		room := &model.Room{
			RoomId:helper.GetUuid(),
			RoomType:define.ROOM_TYPE_SINGLE,
			RoomStatus:define.ROOM_STATUS_OK,
			RoomHostUserId:userToken.UserId,
			LastUserId:userToken.UserId,
			LastMessageId:msg.MessageId,
			LastMessageUpdatedAt:currentTime,
			CreateAt:currentTime,
		}
		err = model.RoomInsertOne(room)
		if err != nil {
			log.Printf("[DB ERROR] %v\n", err)
			helper.FailResponse(ctx, "系统错误")
			return
		}

		// 建立私聊房间关系
		userRoom := [] interface{} {
			&model.UserRoom{
				UserId:userToken.UserId,
				RoomId:room.RoomId,
				RoomType:room.RoomType,
				CreateAt:currentTime,
			},
			&model.UserRoom{
				UserId:toUserId,
				RoomId:room.RoomId,
				RoomType:room.RoomType,
				CreateAt:currentTime,
			},
		}
		err = model.UserRoomInsertMany(userRoom)
		if err != nil {
			log.Printf("[DB ERROR] %v\n", err)
			helper.FailResponse(ctx, "系统错误")
			return
		}

		singleRoomId = room.RoomId
	} else {
		// 更新房间最后一条消息信息
		err := model.UpdateRoomLastMessageByRoomId(singleRoomId, userToken.UserId,msg.MessageId,currentTime)
		if err != nil {
			log.Printf("[DB ERROR] %v\n", err)
		}
	}

	// 记录聊天消息
	userMessage := &model.UserMessage{
		RoomId:singleRoomId,
		MessageId:msg.MessageId,
		SendUserId:userToken.UserId,
		SendStatus:define.MESSAGE_STATUS_OK,
		CreateAt:currentTime,
	}
	err = model.UserMessageInsertOne(userMessage)
	if err != nil {
		log.Printf("[DB ERROR] %v\n", err)
	}

	// 记录用户参与过聊天的房间
	userRoomChatLog := &model.UserRoomChatLog {
		UserId:userToken.UserId,
		RoomId:singleRoomId,
		RoomType:define.ROOM_TYPE_SINGLE,
		LastUpdatedAt:currentTime,
	}
	err = model.UserRoomChatLogInsertOrUpdateOne(userRoomChatLog)
	if err != nil {
		log.Printf("[DB ERROR] 新建聊天列表失败, err: %v\n", err)
		helper.FailResponse(ctx, "系统错误")
		return
	}

	// 推送消息
	messageId := msg.MessageId
	messageType := msg.MessageType
	sendMessageData := define.ResponseMessage{
		MessageId:msg.MessageId,
		MessageType:msg.MessageType,
		MessageData:msg.MessageData,
	}
	messageData, _ := json.Marshal(sendMessageData)
	toUserIdList, err := json.Marshal([]string{toUserId})
	if err != nil {
		log.Printf("[JSON ERROR] 转化json错误, err: %v\n", err)
	}

	// 获取用户ws所在的服务器ip:port
	toAddress, err := cache.Redis.Get(context.Background(), define.REDIS_STRING_USER_WEBSOCKET_CONNECT + toUserId).Result()
	if err != nil {
		log.Printf("[REDIS ERROR] redis错误, err: %v\n", err)
	}
	// port,_ := helper.GetApiPort()
	port,_ := helper.GetWsPort()
	// 请求推送消息接口
	body := "to_user_id="+string(toUserIdList)+"&message_id="+messageId+"&message_type="+strconv.Itoa(int(messageType))+"&message_data="+string(messageData)
	err = helper.SendPost("http://" + toAddress + ":" + port + "/user/push_message", body)
	if err != nil {
		log.Printf("[DB ERROR] 消息推送失败, err: %v\n", err)
		helper.FailResponse(ctx, err.Error())
	}

	helper.SucResponse(ctx, "发送成功", map[string]string{})
}

func RoomMessage(ctx *gin.Context) {
	userToken := ctx.MustGet("user_token").(*helper.UserToken)
	message := ctx.PostForm("message")
	roomId := ctx.PostForm("room_id")
	currentTime := time.Now().Unix()

	// 查询用户是否处于该房间
	_, err := model.GetUserRoomByRoomIdUserId(roomId, userToken.UserId)
	if err == mongo.ErrNoDocuments {
		helper.FailResponse(ctx, "你不在房间")
		return
	}

	if err != nil {
		log.Printf("[DB ERROR] %v\n", err)
		helper.FailResponse(ctx, "系统错误")
		return
	}

	// 添加消息
	msg := &model.Message{
		MessageId:helper.GetUuid(),
		MessageData:message,
		MessageType:define.MESSAGE_TYPE_TEXT,
		MessageStatus:define.MESSAGE_STATUS_OK,
		CreateAt:currentTime,
	}
	err = model.MessageInsertOne(msg)
	if err != nil {
		log.Printf("[DB ERROR] 新建消息失败, err: %v\n", err)
		helper.FailResponse(ctx, "系统错误")
		return
	}

	// 添加房间消息
	userMessage := &model.UserMessage{
		RoomId:roomId,
		MessageId:msg.MessageId,
		SendUserId:userToken.UserId,
		SendStatus:define.MESSAGE_STATUS_OK,
		CreateAt:currentTime,
	}
	err = model.UserMessageInsertOne(userMessage)
	if err != nil {
		log.Printf("[DB ERROR] 新建房间消息失败, err: %v\n", err)
		helper.FailResponse(ctx, "系统错误")
		return
	}

	// 记录用户参与过聊天的房间
	userRoomChatLog := &model.UserRoomChatLog {
		UserId:userToken.UserId,
		RoomId:roomId,
		RoomType:define.ROOM_TYPE_MANY,
		LastUpdatedAt:currentTime,
	}
	err = model.UserRoomChatLogInsertOrUpdateOne(userRoomChatLog)
	if err != nil {
		log.Printf("[DB ERROR] 新建聊天列表失败, err: %v\n", err)
		helper.FailResponse(ctx, "系统错误")
		return
	}

	// 获取房间所有用户
	sendUserIdList := make([]string, 0)
	userRoomList, err := model.GetUserRoomListByRoomId(roomId)
	if err != nil {
		log.Printf("[DB ERROR] 获取用户列表失败, err: %v\n", err)
		helper.FailResponse(ctx, "系统错误")
		return
	}

	// 获取所有的用户id
	for _, v := range userRoomList {
		// 不需要推送给自己
		if v.UserId == userToken.UserId {
			continue
		}
		sendUserIdList = append(sendUserIdList, v.UserId)
	}
	// 元素去重
	uniqueUserIdList := helper.RemoveRepeatedElement(sendUserIdList)
	// 转json
	toUserIdList, _ := json.Marshal(uniqueUserIdList)

	// 发送信息
	messageType := msg.MessageType
	messageId := msg.MessageId
	sendMessageData := define.ResponseMessage{
		MessageId:msg.MessageId,
		MessageType:msg.MessageType,
		MessageData:msg.MessageData,
	}
	messageData, _ := json.Marshal(sendMessageData)

	// 获取所有websocket连接ip
	addressList, err := helper.GetWsAddress()
	if err != nil {
		log.Printf("[CONFIG ERROR] 获取wc地址错误，err:%v\n", err)
	}

	// 获取端口
	// port,err := helper.GetApiPort()
	port,err := helper.GetWsPort()
	if err != nil {
		log.Printf("[CONFIG ERROR] 获取wc端口错误，err:%v\n", err)
	}

	// 群聊的话使用广播的方式发送数据
	body := "to_user_id="+string(toUserIdList)+"&message_id="+messageId+"&message_type="+strconv.Itoa(int(messageType))+"&message_data="+string(messageData)
	wg := &sync.WaitGroup{}
	wg.Add(len(addressList))
	log.Println("群发消息地址", addressList)
	for _, address := range addressList {
		go func(address, body string, wg *sync.WaitGroup) {
			err = helper.SendPost("http://"+address+":"+port+"/user/push_message", body)
			if err != nil {
				log.Printf("[PUSH MESSAGE ERROR] %v\n", err)
			}
			log.Println(address + "发送完毕")
			wg.Done()
		}(address, body, wg)
	}

	wg.Wait()

	helper.SucResponse(ctx, "消息发送成功", make(map[string]string))
}