package test

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"simpleGinIm/model"
	"testing"
	"time"
)

func TestInsertMany(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().
		ApplyURI("mongodb://192.168.78.135").SetAuth(options.Credential{
		Username:"golang",
		Password:"golang",
	}))

	if err != nil {
		log.Println("Mongo Connection err", err)
	}

	db := client.Database("simple_im")

	// 建立私聊房间关系
	userRoom := [] interface{} {
		&model.UserRoom{
			UserId:"1",
			RoomId:"222",
			RoomType:1,
			CreateAt:12321,
		},
		&model.UserRoom{
			UserId:"2",
			RoomId:"222",
			RoomType:1,
			CreateAt:12321,
		},
	}

	_, err = db.Collection("user_room").InsertMany(context.Background(), userRoom, options.InsertMany().SetOrdered(false))
	if err != nil {
		t.Fatal(err)
	}
}
