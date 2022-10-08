package model

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
)

/*
{
  "user_id": "系统接入的用户id",
  "login_status": "登录状态【0离线1在线】",
  "last_login_time": "最后一次登录时间",
  "last_login_out_time": "最后一个退出时间",
  "user_status": "用户状态【1正常-1删除】",
  "created_at": "创建时间"
}
*/
type User struct {
	UserId string `bson:"user_id"`
	LoginStatus int64 `bson:"login_status"`
	LastLoginTime int64 `bson:"last_login_time"`
	LastLoginOutTime int64 `bson:"last_login_out_time"`
	UserStatus int64 `bson:"user_status"`
	CreateAt int64 `bson:"created_at"`
}

func(User) CollectionName() string {
	return "user"
}

func GetUserByUserId(userId string) (*User, error) {
	user := &User{}

	err := InitMongo().Collection(User{}.CollectionName()).
		FindOne(context.Background(), bson.M{"user_id":userId}).
		Decode(user)

	return user, err
}

func UserInsertOne(user *User) error {
	_, err := Mongo.Collection(User{}.CollectionName()).
		InsertOne(context.Background(), user)
	return err
}

func UpdateUserLoginStatusByUserId(userId string, loginStatus, lastLoginTime int64) error {
	filter := bson.D{{"user_id", userId}}
	update := bson.D{{"$set", bson.D{
		{"login_status", loginStatus},
		{"last_login_time", lastLoginTime},
	}}}

	_, err := Mongo.Collection(User{}.CollectionName()).
		UpdateOne(context.TODO(), filter, update)

	return err
}

func UpdateUserLoginStatusByUserIdList(userId []string, loginStatus, lastLoginTime int64) error {
	filter := bson.D{
		{"user_id", bson.D{
			{"$in",userId},
		}},
	}
	update := bson.D{{"$set", bson.D{
		{"login_status", loginStatus},
		{"last_login_time", lastLoginTime},
	}}}

	_, err := Mongo.Collection(User{}.CollectionName()).
		UpdateOne(context.TODO(), filter, update)

	return err
}

func UpdateUserLoginOutStatusByUserId(userId string, loginStatus, lastLoginOutTime int64) error {
	filter := bson.D{{"user_id", userId}}
	update := bson.D{{"$set", bson.D{
		{"login_status", loginStatus},
		{"last_login_out_time", lastLoginOutTime},
	}}}

	_, err := Mongo.Collection(User{}.CollectionName()).
		UpdateOne(context.TODO(), filter, update)

	return err
}

func UpdateUserLoginOutStatusByUserIdList(userId []string, loginStatus, lastLoginOutTime int64) error {
	filter := bson.D{
		{"user_id", bson.D{
			{"$in",userId},
		}},
	}
	update := bson.D{{"$set", bson.D{
		{"login_status", loginStatus},
		{"last_login_out_time", lastLoginOutTime},
	}}}

	_, err := Mongo.Collection(User{}.CollectionName()).
		UpdateOne(context.TODO(), filter, update)

	return err
}
