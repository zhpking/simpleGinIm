package test

import (
	"gopkg.in/ini.v1"
	"simpleGinIm/define"
	"testing"
	"fmt"
)

func TestGetConfig(t *testing.T) {
	path := define.GetDbConfigPath()
	cfg, err := ini.Load(path)
	if err != nil {
		t.Fatal(err)
	}

	// 获取mysql分区的key
	fmt.Println(cfg.Section("mongo").Key("database").String()) // 将结果转为string
	// t.Log(cfg.Section("mongo").Key("port").Int())    // 将结果转为int
}
