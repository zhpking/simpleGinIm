package model

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
{
  "user_id": "用户id",
  "room_id": "房间id",
  "created_at": "创建时间"
  "updated_at": "更新时间"
}
*/

type UserRoomChatLog struct {
	UserId string `bson:"user_id"`
	RoomId string `bson:"room_id"`
	RoomType int64 `bson:"room_type"`
	LastUpdatedAt int64 `bson:"last_updated_at"`
}

func (UserRoomChatLog) CollectionName() string {
	return "user_room_chat_log"
}

func UserRoomChatLogInsertOrUpdateOne(log *UserRoomChatLog) error {
	opts := options.Update().SetUpsert(true)
	filter := bson.D{
		{"user_id", log.UserId},
		{"room_id",log.RoomId},
	}
	update := bson.D{
		{"$set", bson.D {
			{"last_updated_at", log.LastUpdatedAt},
			{"room_type", log.RoomType},
		},
		},
	}

	_, err := Mongo.Collection(UserRoomChatLog{}.CollectionName()).
		UpdateOne(context.TODO(), filter, update, opts)

	return err
}

func GetUserRoomChatLogByUserId(userId string, roomType, page, pageSize int64) ([]*UserRoomChatLog, error) {
	urList := make([]*UserRoomChatLog, 0)

	limit := pageSize
	skip := (page - 1) * pageSize
	opts := &options.FindOptions{}
	opts = opts.SetLimit(limit).SetSkip(skip).SetSort(bson.D{bson.E{"last_updated_at", -1}})

	filter := bson.D{{"user_id", userId}}
	if roomType != 0 {
		filter = bson.D{{"user_id",userId}, {"room_type", roomType}}
	}

	cur, err := Mongo.Collection(UserRoomChatLog{}.CollectionName()).
		Find(context.Background(), filter, opts)

	if err != nil {
		return urList, err
	}

	for cur.Next(context.Background()) {
		ur := &UserRoomChatLog{}
		err = cur.Decode(ur)
		if err != nil {
			return urList, err
		}

		urList = append(urList, ur)
	}

	return urList, nil
}
