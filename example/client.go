package example

import (
	"bufio"
	"fmt"
	"net"
	"simpleGinIm/connect/handler"
	"simpleGinIm/helper"
)

func TcpClient() {
	port, _ := helper.GetTcpPort()
	addressList, _ := helper.GetTcpAddress()
	conn, err := net.Dial("tcp", addressList[0] + ":" + port)
	if err != nil {
		fmt.Println("dial failed, err", err)
		return
	}
	defer conn.Close()

	go func(c net.Conn) {
		reader := bufio.NewReader(conn)
		for {
			msg, err := handler.Decode(reader)
			if err != nil {
				fmt.Printf("[TCP CLIENT ERROR] err:%v\n", err.Error())
			}
			fmt.Println(string(msg))
		}
	}(conn)

	for {
		fmt.Println("请输入token：")
		msg := ""
		fmt.Scanln(&msg)
		data, err := handler.Encode([]byte(msg))
		if err != nil {
			fmt.Println("encode msg failed, err:", err)
			return
		}

		/*
		for i := 0; i < 20; i ++ {
			// 测试并发投递，看是否会出现粘包
			go conn.Write(data)
		}
		*/

		_, err = conn.Write(data)
		if err != nil {
			fmt.Println("send msg failed, err:", err)
		}
	}

	/*
	for i := 0; i < 20; i++ {
		msg := `Hello, Hello. How are you?`
		data, err := service.Encode(msg)
		if err != nil {
			fmt.Println("encode msg failed, err:", err)
			return
		}
		conn.Write(data)
	}
	*/
}
