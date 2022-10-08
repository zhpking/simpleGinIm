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

	port := cfg.Section("api").Key("port").String()
	return port, nil
}

func GetTcpPort() (string, error) {
	path := define.GetSysConfigPath()
	cfg, err := ini.Load(path)
	if err != nil {
		log.Printf("[SYS CONFIG ERROR] %v\n", err)
		return "", err
	}

	port := cfg.Section("tcp").Key("port").String()
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

func GetTcpAddress() ([]string, error) {
	path := define.GetSysConfigPath()
	cfg, err := ini.Load(path)
	if err != nil {
		log.Printf("[SYS CONFIG ERROR] %v\n", err)
		return []string{}, err
	}

	address := cfg.Section("tcp").Key("address").String()
	addressList := strings.Split(address, ",")

	return addressList, nil
}