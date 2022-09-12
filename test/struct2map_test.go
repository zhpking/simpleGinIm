package test

import (
	"fmt"
	"github.com/fatih/structs"
	"simpleGinIm/helper"
	"simpleGinIm/model"
	"testing"
)

// https://github.com/fatih/structs
func TestStruct(t *testing.T) {
	urList := []*model.UserRoom{
		{
			UserId:   "1",
			RoomId:   "abc1",
			RoomType: 1,
			CreateAt: 123123123,
		},
		{
			UserId:   "2",
			RoomId:   "abc2",
			RoomType: 1,
			CreateAt: 123123123,
		},
		{
			UserId:   "3",
			RoomId:   "abc3",
			RoomType: 1,
			CreateAt: 123123123,
		},
	}

	ret := map[string]interface{}{}
	for _, val := range urList {
		s := structs.New(val)
		value := s.Field("UserId").Value().(string)
		ret[value] = val
	}

	fmt.Println(ret)
}

func TestCombineStructList(t *testing.T) {
	uesrMessageList := []interface{}{
		&model.UserMessage{
			RoomId:     "31373334613264352d346161612d343034622d383966362d373134666136376338343933",
			MessageId:  "35373534303236342d643664332d346663612d626532352d623766393331333263646237",
			SendUserId: "10000200275",
			SendStatus: 1,
			CreateAt:   1662877010,
		},
		&model.UserMessage{
			RoomId:     "31373334613264352d346161612d343034622d383966362d373134666136376338343933",
			MessageId:  "37663662376464392d383466362d346466322d393539352d386435373730326634623865",
			SendUserId: "10000200275",
			SendStatus: 1,
			CreateAt:   1662876992,
		},
	}

	messageList := []interface{} {
		&model.Message{
			MessageId:   "35373534303236342d643664332d346663612d626532352d623766393331333263646237",
			MessageData: "排第一",
			MessageStatus:1,
			CreateAt:1662876992,
		},
		&model.Message{
			MessageId:   "37663662376464392d383466362d346466322d393539352d386435373730326634623865",
			MessageData: "排第二",
			MessageStatus:1,
			CreateAt:1662876992,
		},
	}

	newMessageList := helper.Struct2MapByKey("messageId", messageList)
	fmt.Println(newMessageList)
	// helper.combineStructList(uesrMessageList)
}
