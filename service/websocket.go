package service

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"simpleGinIm/define"
	"simpleGinIm/helper"
	"simpleGinIm/model"
	"time"
)

var upgrader websocket.Upgrader = websocket.Upgrader{}
var wc map[string]*websocket.Conn = make(map[string]*websocket.Conn)

type responseMessage struct {
	MessageId string `json:"message_id"`
	MessageType int64 `json:"message_type"`
	MessageData string `json:"message_data"`
}

func Connect(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		helper.FailResponse(ctx, "系统异常")
		conn.Close()
		return
	}

	ut := ctx.MustGet("user_token").(*helper.UserToken)
	wc[ut.UserId] = conn

	// helper.SucResponse(ctx, "系统异常", make(map[string]interface{}))
}

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
	conn, ok := wc[toUserId]
	if !ok {
		// 对方不在线
		helper.SucResponse(ctx, "发送成功", make(map[string]interface{}))
		return
	}

	sendMessageData := responseMessage{
		MessageId:msg.MessageId,
		MessageType:msg.MessageType,
		MessageData:msg.MessageData,
	}
	responseData, _ := json.Marshal(sendMessageData)

	// err = conn.WriteMessage(websocket.TextMessage, msg.MessageData)
	err = conn.WriteMessage(websocket.TextMessage, responseData)
	if err != nil {
		log.Printf("消息发送失败,error: %v\n", err)
		helper.FailResponse(ctx, "发送失败")
		return
	}
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
		sendUserIdList = append(sendUserIdList, v.UserId)
	}

	sendMessageData := responseMessage{
		MessageId:msg.MessageId,
		MessageType:msg.MessageType,
		MessageData:msg.MessageData,
	}
	responseData, _ := json.Marshal(sendMessageData)

	// 发送信息
	for _, v := range sendUserIdList {
		// 不需要发给自己
		if userToken.UserId == v {
			continue
		}

		if conn, ok := wc[v]; ok {
			// 推送消息
			// err = conn.WriteMessage(websocket.TextMessage, []byte(message))
			err = conn.WriteMessage(websocket.TextMessage, responseData)
			if err != nil {
				log.Printf("[WEBSOCKET] %v发送消息失败，err:%v\n", v, err)
			}
		}
	}

	helper.SucResponse(ctx, "消息发送成功", make(map[string]string))
}

func RemoveUserConnect(userId string) error {
	err := wc[userId].Close()
	if err != nil {
		return err
	}
	delete(wc, userId)
	return nil
}
