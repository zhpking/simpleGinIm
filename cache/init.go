package cache

import (
	"github.com/go-redis/redis/v8"
	"gopkg.in/ini.v1"
	"log"
	"simpleGinIm/define"
)

var Redis = InitRedis()

func InitRedis() *redis.Client {
	// 获取redis配置
	path := define.GetDbConfigPath()
	cfg, err := ini.Load(path)
	if err != nil {
		log.Printf("[DB CONFIG ERROR] %v\n", err)
		return nil
	}
	// 获取mongo分区的key
	address := cfg.Section("redis").Key("address").String() // 将结果转为string
	return redis.NewClient(&redis.Options{
		Addr:address,
	})
}
