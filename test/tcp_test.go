package test

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"
	"testing"
	"time"
)

// https://www.liwenzhou.com/posts/Go/15_socket/
func TestTcpService(t *testing.T) {
	// 启动服务端
	go tcpService ()
	time.Sleep(3 * time.Second)

	// 客户端代码
	conn, err := net.Dial("tcp", "127.0.0.1:30000")
	if err != nil {
		log.Println("dial failed, err", err)
		return
	}
	defer conn.Close()
	for i := 0; i < 20; i++ {
		msg := `Hello Test`
		data, err := Encode(msg)
		if err != nil {
			log.Println("encode msg failed, err:", err)
			return
		}
		conn.Write(data)
	}
}

// Encode 将消息编码
// 前4字节（头部）是消息长度，后面是消息内容（body）
func Encode(message string) ([]byte, error) {
	// 读取消息的长度，转换成int32类型（占4个字节）
	var length = int32(len(message))
	var pkg = new(bytes.Buffer)
	// 写入消息头
	err := binary.Write(pkg, binary.LittleEndian, length)
	if err != nil {
		return nil, err
	}
	// 写入消息实体
	err = binary.Write(pkg, binary.LittleEndian, []byte(message))
	if err != nil {
		return nil, err
	}
	return pkg.Bytes(), nil
}

// Decode 解码消息
func Decode(reader *bufio.Reader) (string, error) {
	// 读取消息的长度
	lengthByte, _ := reader.Peek(4) // 读取前4个字节的数据
	lengthBuff := bytes.NewBuffer(lengthByte)
	var length int32
	err := binary.Read(lengthBuff, binary.LittleEndian, &length)
	if err != nil {
		return "", err
	}
	// Buffered返回缓冲中现有的可读取的字节数。
	if int32(reader.Buffered()) < length+4 {
		return "", err
	}

	// 读取真正的消息数据
	pack := make([]byte, int(4+length))
	_, err = reader.Read(pack)
	if err != nil {
		return "", err
	}
	return string(pack[4:]), nil
}

// tcp服务端处理逻辑
func process(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		msg, err := Decode(reader)
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Println("decode msg failed, err:", err)
			return
		}
		log.Println("收到client发来的数据：", msg)
	}
}

func tcpService () {
	listen, err := net.Listen("tcp", "127.0.0.1:30000")
	if err != nil {
		log.Println("listen failed, err:", err)
		return
	}
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println("accept failed, err:", err)
			continue
		}
		go process(conn)
	}
}
