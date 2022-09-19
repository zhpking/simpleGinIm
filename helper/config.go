package helper

import (
	"gopkg.in/ini.v1"
	"log"
	"simpleGinIm/define"
	"strings"
)

func GetWsPort() (string, error) {
	path := define.GetSysConfigPath()
	cfg, err := ini.Load(path)
	if err != nil {
		log.Printf("[SYS CONFIG ERROR] %v\n", err)
		return "", err
	}

	port := cfg.Section("ws").Key("port").String()
	return port, nil
}

func GetApiPort() (string, error) {
	path := define.GetSysConfigPath()
	cfg, err := ini.Load(path)
	if err != nil {
		log.Printf("[SYS CONFIG ERROR] %v\n", err)
		return "", err
	}

	// 获取mongo分区的key
	port := cfg.Section("api").Key("port").String()
	return port, nil
}

func GetWsAddress() ([]string, error) {
	path := define.GetSysConfigPath()
	cfg, err := ini.Load(path)
	if err != nil {
		log.Printf("[SYS CONFIG ERROR] %v\n", err)
		return []string{}, err
	}

	address := cfg.Section("ws").Key("address").String()
	addressList := strings.Split(address, ",")

	return addressList, nil
}

func GetApiAddress() ([]string, error) {
	path := define.GetSysConfigPath()
	cfg, err := ini.Load(path)
	if err != nil {
		log.Printf("[SYS CONFIG ERROR] %v\n", err)
		return []string{}, err
	}

	address := cfg.Section("api").Key("address").String()
	addressList := strings.Split(address, ",")

	return addressList, nil
}