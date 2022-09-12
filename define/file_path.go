package define

import (
	"os"
)


func getCurrentPath() string {
	// 获取的是main.go所在目录
	currentPath, _ := os.Getwd()
	return currentPath
}

func GetDbConfigPath() string {
	modPath := getCurrentPath()
	return modPath + "/config/db_config.ini"
}

func GetSysConfigPath() string {
	modPath := getCurrentPath()
	return modPath + "/config/sys_config.ini"
}
