package service

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"simpleGinIm/cache"
	"simpleGinIm/define"
	"simpleGinIm/helper"
	"simpleGinIm/model"
	"strconv"
	"sync"
	"time"
)

func UserSendMessage(receiveMsg *define.ReceiveMessage) {
	// 获取房间，根据房间类型，判断是单聊还是群聊
	roomId := receiveMsg.RoomId
	userId := receiveMsg.UserId
	currentTime := time.Now().Unix()
	room, err := model.GetRoomByRoomId(roomId)
	if err == mongo.ErrNoDocuments {
		// 房间不存在
		return
	}
	if err != nil {
		log.Printf("[DB ERROR] %v\n", err)
		return
	}

	// 查询用户是否处于该房间
	_, err = model.GetUserRoomByRoomIdUserId(roomId, userId)
	if err == mongo.ErrNoDocuments {
		// 不在房间
		return
	}
	if err != nil {
		log.Printf("[DB ERROR] %v\n", err)
		return
	}

	// 添加消息
	msg := &model.Message{
		MessageId:helper.GetUuid(),
		MessageData:receiveMsg.MessageData,
		MessageType:receiveMsg.MessageType,
		MessageStatus:define.MESSAGE_STATUS_OK,
		CreateAt:currentTime,
	}
	err = model.MessageInsertOne(msg)
	if err != nil {
		log.Printf("[DB ERROR] 新建消息失败, err: %v\n", err)
		return
	}

	if room.RoomType == define.ROOM_TYPE_SINGLE {
		// 私聊
		SingleMessage(receiveMsg, msg.MessageId, currentTime)
	} else {
		// 群聊
		RoomMessage(receiveMsg, msg.MessageId, currentTime)
	}
}

func RoomMessage(receiveMsg *define.ReceiveMessage, messageId string, currentTime int64) {
	userId := receiveMsg.UserId
	roomId := receiveMsg.RoomId
	messageType := receiveMsg.MessageType
	messageData := receiveMsg.MessageData
	// 记录用户参与过聊天的房间
	userRoomChatLog := &model.UserRoomChatLog {
		UserId:userId,
		RoomId:roomId,
		RoomType:define.ROOM_TYPE_MANY,
		LastUpdatedAt:currentTime,
	}
	err := model.UserRoomChatLogInsertOrUpdateOne(userRoomChatLog)
	if err != nil {
		log.Printf("[DB ERROR] 新建聊天列表失败, err: %v\n", err)
		return
	}

	// 获取房间所有用户
	sendUserIdList := make([]string, 0)
	userRoomList, err := model.GetUserRoomListByRoomId(roomId)
	if err != nil {
		log.Printf("[DB ERROR] 获取用户列表失败, err: %v\n", err)
		return
	}

	// 获取所有的用户id
	for _, v := range userRoomList {
		// 不需要推送给自己
		if v.UserId == userId {
			continue
		}
		sendUserIdList = append(sendUserIdList, v.UserId)
	}

	toUserIdList, _ := json.Marshal(sendUserIdList)

	// 获取所有websocket连接
	addressList, err := helper.GetWsAddress()
	if err != nil {
		log.Printf("[CONFIG ERROR] 获取wc地址错误，err:%v\n", err)
		return
	}

	// 获取端口
	port,err := helper.GetWsPort()
	if err != nil {
		log.Printf("[CONFIG ERROR] 获取wc端口错误，err:%v\n", err)
		return
	}

	// 群聊的话使用广播的方式发送数据
	body := "to_user_id="+string(toUserIdList)+"&message_id="+messageId+"&message_type="+strconv.Itoa(int(messageType))+"&message_data="+messageData
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

func SingleMessage(receiveMsg *define.ReceiveMessage, messageId string, currentTime int64) {
	userId := receiveMsg.UserId
	toUserId := receiveMsg.ToUserId
	messageType := receiveMsg.MessageType
	messageData := receiveMsg.MessageData


	// 两人是否拥有房间
	isHasRoom := true

	// 获取单聊房间
	sendUrList, err := model.GetUserRoomListByUserId(userId, define.ROOM_TYPE_SINGLE)
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
			RoomHostUserId:userId,
			LastUserId:userId,
			LastMessageId:messageId,
			LastMessageUpdatedAt:currentTime,
			CreateAt:currentTime,
		}
		err = model.RoomInsertOne(room)
		if err != nil {
			log.Printf("[DB ERROR] %v\n", err)
			return
		}

		// 建立私聊房间关系
		userRoom := [] interface{} {
			&model.UserRoom{
				UserId:userId,
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
			return
		}

		singleRoomId = room.RoomId
	} else {
		// 更新房间最后一条消息信息
		err := model.UpdateRoomLastMessageByRoomId(singleRoomId, userId, messageId, currentTime)
		if err != nil {
			log.Printf("[DB ERROR] %v\n", err)
		}
	}

	// 记录聊天消息
	userMessage := &model.UserMessage{
		RoomId:singleRoomId,
		MessageId:messageId,
		SendUserId:userId,
		SendStatus:define.MESSAGE_STATUS_OK,
		CreateAt:currentTime,
	}
	err = model.UserMessageInsertOne(userMessage)
	if err != nil {
		log.Printf("[DB ERROR] %v\n", err)
	}

	// 记录用户参与过聊天的房间
	userRoomChatLog := &model.UserRoomChatLog {
		UserId:userId,
		RoomId:singleRoomId,
		RoomType:define.ROOM_TYPE_SINGLE,
		LastUpdatedAt:currentTime,
	}
	err = model.UserRoomChatLogInsertOrUpdateOne(userRoomChatLog)
	if err != nil {
		log.Printf("[DB ERROR] 新建聊天列表失败, err: %v\n", err)
		return
	}


	toUserIdList, err := json.Marshal([]string{toUserId})
	if err != nil {
		log.Printf("[JSON ERROR] 转化json错误, err: %v\n", err)
	}

	// 获取用户ws所在的服务器ip:port
	toAddress, err := cache.Redis.Get(context.Background(), define.REDIS_STRING_USER_WEBSOCKET_CONNECT + toUserId).Result()
	if err != nil {
		log.Printf("[REDIS ERROR] redis错误, err: %v\n", err)
	}
	// 请求推送消息接口
	body := "to_user_id="+string(toUserIdList)+"&message_id="+messageId+"&message_type="+strconv.Itoa(int(messageType))+"&message_data="+messageData
	err = helper.SendPost("http://" + toAddress + "/user/push_message", body)
	if err != nil {
		log.Printf("[DB ERROR] 消息推送失败, err: %v\n", err)
		return
	}
}