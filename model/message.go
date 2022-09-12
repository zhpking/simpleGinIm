package model

import (
	"context"
)

/*
{
  "message_id": "消息id",
  "message_data": "消息数据",
  "message_type": "消息类型【1文本消息】",
  "created_at": "创建时间",
  "message_status": "消息状态【1正常】"
}
*/
type Message struct {
	MessageId string `bson:"message_id"`
	MessageData string `bson:"message_data"`
	MessageType int64 `bson:"message_type"`
	MessageStatus int64 `bson:"message_status"`
	CreateAt int64 `bson:"created_at"`
}

func(Message) CollectionName() string {
	return "message"
}

func MessageInsertOne(message *Message) error {
	_, err := Mongo.Collection(Message{}.CollectionName()).InsertOne(context.Background(), message)
	return err
}
