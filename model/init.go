package model

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/ini.v1"
	"log"
	"simpleGinIm/define"
	"time"
)

var Mongo = InitMongo()

func InitMongo() *mongo.Database {
	// 获取mongo配置
	path := define.GetDbConfigPath()
	cfg, err := ini.Load(path)
	if err != nil {
		log.Printf("[DB CONFIG ERROR] %v\n", err)
		return nil
	}

	// 获取mongo分区的key
	address := cfg.Section("mongo").Key("address").String() // 将结果转为string
	username := cfg.Section("mongo").Key("username").String() // 将结果转为string
	password := cfg.Section("mongo").Key("password").String() // 将结果转为string
	database := cfg.Section("mongo").Key("database").String() // 将结果转为string

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().
		ApplyURI("mongodb://" + address).SetAuth(options.Credential{
		Username:username,
		Password:password,
	}))

	if err != nil {
		log.Println("Mongo Connection err", err)
	}

	return client.Database(database)
}
