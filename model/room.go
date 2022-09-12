package model

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
{
  "room_id": "房间id",
  "room_type": "房间类型【1单人2多人】",
  "room_status": "房间状态【1正常-1删除】",
  "last_user_id": "最后一个发消息的用户id",
  "last_message_id": "最后一条消息id",
  "last_message_updated_at": "最后一条消息更新时间",
  "created_at": "创建时间"
}
*/

type Room struct {
	RoomId string `bson:"room_id"`
	RoomType int64 `bson:"room_type"`
	RoomStatus int64 `bson:"room_status"`
	RoomHostUserId string `bson:"room_host_user_id"`
	LastUserId string `bson:"last_user_id"`
	LastMessageId string `bson:"last_message_id"`
	LastMessageUpdatedAt int64 `bson:"last_message_updated_at"`
	CreateAt int64 `bson:"created_at"`
}

func(Room) CollectionName() string {
	return "room"
}

func RoomInsertOne(room *Room) error {
	_ , err := Mongo.Collection(Room{}.CollectionName()).InsertOne(context.Background(), room)
	return err
}

func UpdateRoomLastMessageByRoomId(roomId,LastUserId,LastMessageId string, LastMessageUpdatedAt int64) error {
	filter := bson.M{"room_id":roomId}
	update := bson.D{
		{"$set", bson.D{
			{"last_user_id", LastUserId},
			{"last_message_id", LastMessageId},
			{"last_message_updated_at", LastMessageUpdatedAt},
		},
		},
	}
	_, err := Mongo.Collection(Room{}.CollectionName()).UpdateOne(context.Background(), filter, update)
	return err
}

func GetRoomByRoomId(roomId string) (*Room, error) {
	ur := &Room{}
	err := Mongo.Collection(Room{}.CollectionName()).
		FindOne(context.Background(), bson.M{"room_id":roomId}).
		Decode(ur)

	return ur, err
}

func GetRoomListByRoomId(roomId []string) ([]*Room, error) {
	rList := make([]*Room, 0)
	filter := bson.D{
		{"room_id", bson.D{
			{"$in", roomId},
		},
		},
	}
	opts := (&options.FindOptions{}).SetSort(bson.D{bson.E{"last_message_updated_at", -1}})
	cur, err := Mongo.Collection(Room{}.CollectionName()).
		Find(context.Background(), filter, opts)
	if err != nil {
		return rList, err
	}

	for cur.Next(context.Background()) {
		r := &Room{}
		err = cur.Decode(r)
		if err != nil {
			return rList, err
		}

		rList = append(rList, r)
	}

	return rList, nil
}
