package model

import (
	"go.mongodb.org/mongo-driver/bson"
	"context"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
{
  "user_id": "用户id",
  "room_id": "房间id",
  "created_at": "房间id"
}
*/

type UserRoom struct {
	UserId string `bson:"user_id"`
	RoomId string `bson:"room_id"`
	RoomType int64 `bson:"room_type"`
	CreateAt int64 `bson:"created_at"`
}

func(UserRoom) CollectionName() string {
	return "user_room"
}

func GetUserRoomListByUserId(userId string, roomType int64) ([]*UserRoom, error) {
	userRoomList := make([]*UserRoom, 0)

	filter := bson.D{{"user_id", userId}, {"room_type", roomType}}
	if roomType == 0 {
		filter = bson.D{{"user_id", userId}}
	}

	cur, err := Mongo.Collection(UserRoom{}.CollectionName()).
		Find(context.Background(), filter)

	if err != nil {
		return userRoomList, err
	}

	for cur.Next(context.Background()) {
		ur := &UserRoom{}
		err = cur.Decode(ur)
		if err != nil {
			return userRoomList, err
		}

		userRoomList = append(userRoomList, ur)
	}

	return userRoomList, nil
}

func UserRoomInsertOne(userRoom *UserRoom) error {
	_, err := Mongo.Collection(UserRoom{}.CollectionName()).InsertOne(context.Background(), userRoom)
	return err
}

func UserRoomInsertMany(userRoom []interface{}) error {
	_, err := Mongo.Collection(UserRoom{}.CollectionName()).InsertMany(context.Background(), userRoom, options.InsertMany().SetOrdered(false))
	return err
}

func RemoveUserRoomByRoomIdUserId(roomId, userId string) error {
	filter := bson.D{{"room_id", roomId}, {"user_id", userId}}
	_, err := Mongo.Collection(UserRoom{}.CollectionName()).DeleteOne(context.Background(), filter)
	return err
}

func GetUserRoomByRoomIdUserId(roomId, userId string) (*UserRoom, error) {
	ur := &UserRoom{}
	filter := bson.D{{"room_id", roomId}, {"user_id", userId}}
	err := Mongo.Collection(UserRoom{}.CollectionName()).
		FindOne(context.Background(), filter).
		Decode(ur)
	return ur, err
}

func GetUserRoomListByRoomId(roomId string) ([]*UserRoom, error) {
	userRoomList := make([]*UserRoom, 0)

	filter := bson.D{{"room_id", roomId}}
	cur, err := Mongo.Collection(UserRoom{}.CollectionName()).
		Find(context.Background(), filter)

	if err != nil {
		return userRoomList, err
	}

	for cur.Next(context.Background()) {
		ur := &UserRoom{}
		err = cur.Decode(ur)
		if err != nil {
			return userRoomList, err
		}

		userRoomList = append(userRoomList, ur)
	}

	return userRoomList, nil
}
