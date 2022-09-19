package test

import (
	"log"
	"simpleGinIm/helper"
	"testing"
)

func TestGetLocalIp(t *testing.T) {
	ipList := helper.GetLocalIP()
	log.Println(ipList)
}
