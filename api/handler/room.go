package handler

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"simpleGinIm/define"
	"simpleGinIm/helper"
	"simpleGinIm/model"
	"strconv"
	"time"
)

func CreateRoom(ctx *gin.Context) {
	userToken := ctx.MustGet("user_token").(*helper.UserToken)
	roomTypeStr := ctx.DefaultPostForm("room_type", strconv.Itoa(define.ROOM_TYPE_MANY))
	roomType, _ := strconv.ParseInt(roomTypeStr, 10, 64)
	currentTime := time.Now().Unix()

	room := &model.Room{
		RoomId:helper.GetUuid(),
		RoomType:roomType,
		RoomStatus:define.ROOM_STATUS_OK,
		RoomHostUserId:userToken.UserId,
		LastUserId:"",
		LastMessageId:"",
		LastMessageUpdatedAt:0,
		CreateAt:currentTime,
	}

	err := model.RoomInsertOne(room)
	if err != nil {
		log.Println("[DB ERROR] %v\n", err)
		helper.FailResponse(ctx, "系统错误")
		return
	}

	// 进入房间
	userRoom := &model.UserRoom{
		UserId:userToken.UserId,
		RoomId:room.RoomId,
		RoomType:room.RoomType,
		CreateAt:currentTime,
	}
	err = model.UserRoomInsertOne(userRoom)
	if err != nil {
		log.Println("[DB ERROR] %v\n", err)
		helper.FailResponse(ctx, "创建房间成功，请重新进入房间")
		return
	}

	helper.SucResponse(ctx, "创建成功", map[string]string{"roomId":room.RoomId})
}

func EnterRoom(ctx *gin.Context) {
	userToken := ctx.MustGet("user_token").(*helper.UserToken)
	roomId := ctx.PostForm("room_id")
	currentTime := time.Now().Unix()

	// 获取房间信息
	room, err := model.GetRoomByRoomId(roomId)
	if err == mongo.ErrNoDocuments {
		helper.FailResponse(ctx, "房间不存在")
		return
	}
	if err != nil {
		log.Printf("[DB ERROR] %v\n", err)
		helper.FailResponse(ctx, "系统错误")
		return
	}

	userRoom := &model.UserRoom{
		UserId:userToken.UserId,
		RoomId:roomId,
		RoomType:room.RoomType,
		CreateAt:currentTime,
	}

	err = model.UserRoomInsertOne(userRoom)
	if err != nil {
		log.Printf("[DB ERROR] %v\n", err)
		helper.FailResponse(ctx, "进入房间失败")
		return
	}

	helper.SucResponse(ctx, "进入房间成功", make(map[string]string))
}

func InviteUserEnterRoom(ctx *gin.Context) {
	userToken := ctx.MustGet("user_token").(*helper.UserToken)
	inviteUserId := ctx.PostForm("invite_user_id")
	roomId := ctx.PostForm("room_id")
	currentTime := time.Now().Unix()

	// 只有房主能邀请人
	room, err := model.GetRoomByRoomId(roomId)
	if err != nil {
		log.Printf("[DB ERROR] %v\n", err)
		helper.FailResponse(ctx, "获取房间信息失败")
		return
	}

	if room.RoomHostUserId != userToken.UserId {
		helper.FailResponse(ctx, "只有房主能邀请人")
		return
	}

	userRoom := &model.UserRoom{
		UserId:inviteUserId,
		RoomId:roomId,
		RoomType:room.RoomType,
		CreateAt:currentTime,
	}

	err = model.UserRoomInsertOne(userRoom)
	if err != nil {
		log.Printf("[DB ERROR] %v\n", err)
		helper.FailResponse(ctx, "邀请进入房间失败")
		return
	}

	helper.SucResponse(ctx, "邀请进入房间成功", make(map[string]string))
}

func ExitRoom(ctx *gin.Context) {
	userToken := ctx.MustGet("user_token").(*helper.UserToken)
	roomId := ctx.PostForm("room_id")

	err := model.RemoveUserRoomByRoomIdUserId(roomId, userToken.UserId)
	if err != nil {
		log.Printf("[DB ERROR] %v\v", err)
		helper.FailResponse(ctx, "离开房间失败")
		return
	}

	helper.SucResponse(ctx, "离开房间成功", make(map[string]string))
}

func KickOutRoom(ctx *gin.Context) {
	userToken := ctx.MustGet("user_token").(*helper.UserToken)
	userId := ctx.PostForm("user_id")
	roomId := ctx.PostForm("room_id")

	// 只有房主能踢人
	room, err := model.GetRoomByRoomId(roomId)
	if err != nil {
		log.Printf("[DB ERROR] %v\n", err)
		helper.FailResponse(ctx, "获取房间信息失败")
		return
	}

	if room.RoomHostUserId != userToken.UserId {
		helper.FailResponse(ctx, "只有房主能踢人")
		return
	}

	err = model.RemoveUserRoomByRoomIdUserId(roomId, userId)
	if err != nil {
		log.Printf("[DB ERROR] %v\v", err)
		helper.FailResponse(ctx, "踢出房间失败")
		return
	}

	helper.SucResponse(ctx, "踢出房间成功", make(map[string]string))
}

func GetRoomList(ctx *gin.Context) {
	userToken := ctx.MustGet("user_token").(*helper.UserToken)
	roomTypeStr := ctx.DefaultQuery("room_type", "0")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")
	roomType, _ := strconv.ParseInt(roomTypeStr, 10, 64)
	page, _ := strconv.ParseInt(pageStr, 10, 64)
	pageSize, _ := strconv.ParseInt(pageSizeStr, 10, 64)

	urList , err := model.GetUserRoomChatLogByUserId(userToken.UserId, roomType, page, pageSize)
	if err == mongo.ErrNoDocuments {
		helper.SucResponse(ctx, "获取成功", []int{})
		return
	}

	if err != nil {
		log.Printf("[DB ERROR] %v\n", err)
		helper.FailResponse(ctx, "系统错误")
		return
	}

	// 获取房间id
	roomIdList := make([]string, 0)
	for _, v := range urList {
		roomIdList = append(roomIdList, v.RoomId)
	}

	// 获取房间信息
	roomList, err := model.GetRoomListByRoomId(roomIdList)
	if err != nil {
		log.Printf("[DB ERROR] %v\n", err)
		helper.FailResponse(ctx, "系统错误")
		return
	}

	helper.SucResponse(ctx, "获取成功", roomList)
}

func GetRoomMessageList(ctx *gin.Context) {
	// userToken := ctx.MustGet("user_token").(*helper.UserToken)
	roomId := ctx.Query("room_id")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")
	page, _ := strconv.ParseInt(pageStr, 10, 64)
	pageSize, _ := strconv.ParseInt(pageSizeStr, 10, 64)

	umList, err := model.GetUserMessageByRoomId(roomId, page, pageSize)
	if err == mongo.ErrNoDocuments {
		helper.SucResponse(ctx, "获取成功", []int{})
		return
	}

	if err != nil {
		log.Printf("[DB ERROR] %v\n", err)
		helper.FailResponse(ctx, "系统错误")
		return
	}

	helper.SucResponse(ctx, "获取成功", umList)
}
