package model

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
{
  "room_id": "房间id",
  "message_id": "消息id",
  "send_user_id": "发送人id",
  "created_at": "创建时间",
  "send_status": "发送状态【1发送成功-1撤回】"
}
*/

type UserMessage struct {
	RoomId string `bson:"room_id"`
	MessageId string `bson:"message_id"`
	SendUserId string `bson:"send_user_id"`
	SendStatus int64 `bson:"send_status"`
	CreateAt int64 `bson:"created_at"`
}

func(UserMessage) CollectionName() string {
	return "user_message"
}

func UserMessageInsertOne(message *UserMessage) error {
	_, err := Mongo.Collection(UserMessage{}.CollectionName()).InsertOne(context.Background(), message)
	return err
}

func GetUserMessageByRoomId(roomId string, page, pageSize int64) ([]*UserMessage, error) {
	userMessageList := make([]*UserMessage, 0)

	limit := pageSize
	skip := (page - 1) * pageSize
	// opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.D{{"created_at", 1}})
	opts := &options.FindOptions{}
	opts = opts.SetLimit(limit).SetSkip(skip).SetSort(bson.D{bson.E{"created_at", -1}})
	filter := bson.M{"room_id":roomId}
	cur, err := Mongo.Collection(UserMessage{}.CollectionName()).
		Find(context.Background(), filter, opts)

	if err != nil {
		return userMessageList, err
	}

	for cur.Next(context.Background()) {
		um := &UserMessage{}
		err = cur.Decode(um)
		if err != nil {
			return userMessageList, err
		}

		userMessageList = append(userMessageList, um)
	}

	return userMessageList, err
}

