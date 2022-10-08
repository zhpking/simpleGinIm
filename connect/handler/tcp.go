package handler

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"
	"simpleGinIm/cache"
	"simpleGinIm/define"
	"simpleGinIm/helper"
	"sync"
	"time"
)

// 关联用户socket
// var ts map[string]*TcpConn = make(map[string]*TcpConn)

// 连接之后，临时存储的socket，此时还没关联用户
// var connectTs map[*TcpConn]int64 = make(map[*TcpConn]int64)

var tcpConnManager TcpConnManager = TcpConnManager{
	ts:make(map[string]*TcpConn),
	connectTs:make(map[*TcpConn]int64),
	lock:sync.RWMutex{},
}

type TcpConnManager struct {
	ts map[string]*TcpConn // 关联用户socket
	connectTs map[*TcpConn]int64 // 连接之后，临时存储的socket，此时还没关联用户
	lock sync.RWMutex
}

type TcpConn struct {
	conn net.Conn
	lock sync.Mutex
	expireTime int64
}

func TcpConnect() {
	port, _ := helper.GetTcpPort()
	listen, err := net.Listen("tcp", ":" + port)
	if err != nil {
		log.Println("[TCP CONNECT ERROR] %v\n", err)
		return
	}
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println("[TCP ACCEPT ERROR] %v\n", err)
			continue
		}
		tcpConn := &TcpConn{
			conn:conn,
			lock:sync.Mutex{},
		}
		tcpConnManager.connectTs[tcpConn] = time.Now().Unix() + 3 // 3s过期
		go process(tcpConn)
	}
}

// tcp服务端处理逻辑
func process(tcpConn *TcpConn) {
	conn := tcpConn.conn
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		msg, err := Decode(reader)
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Printf("[TCP DECODE ERROR] decode msg failed, err:%v\n", err)
			return
		}

		// 因为tcp不像是websocket能通过劫持http升级为websocket
		// 所以需要跟客户端约定，连接成功后，第一条消息是验证token合法性的消息
		// 只有通过校验后，tcp连接才合法
		// 否则3s后连接会自动断开
		userToken, err := helper.AnalyseToken(string(msg))
		if err != nil {
			delete(tcpConnManager.connectTs, tcpConn)
			return
		}

		// 校验过期时间
		userId := userToken.UserId
		if userToken.LoginExpire != 0 && userToken.LoginExpire < time.Now().Unix() {
			delete(tcpConnManager.connectTs, tcpConn)
			// 退出登录队列
			cache.SetUserLoginOutList(userId)
			return
		}

		// 合法的话，就绑定连接
		tcpConn.expireTime = userToken.LoginExpire
		ValidTcpConnect(tcpConn, userId)


		ipList := helper.GetLocalIP()
		// 关联用户登录服务器ip
		cache.SetConnection2User(userId, ipList[0])
		// 登录消息投递
		cache.SetUserLoginList(userId)
		// log.Println("userId：" + userId + "tcp连接校验成功")
	}
}

// tcp心跳检测
func CheckTcpConn() {
	pingPeriod := define.WEBSOCKET_PING_PERIOD * time.Second
	ticker := time.NewTicker(pingPeriod)
	ts := tcpConnManager.ts

	for {
		select {
		case <-ticker.C:
			log.Println("开始心跳检测")
			currentTime := time.Now().Unix()
			if len(ts) == 0 {
				return
			}

			wg := &sync.WaitGroup{}
			wg.Add(len(ts))
			for userId, c := range ts {
				// 超时登录&心跳检测
				go func(tsCon *TcpConn, wg *sync.WaitGroup) {
					if tsCon.expireTime != 0 && tsCon.expireTime < currentTime {
						// 登录超时，退出登录
						log.Println("tcp心跳检测:" + userId + "退出登录")
						// 删除用户长连接
						RemoveUserTcpConnect(userId)
						// 退出登录队列
						cache.SetUserLoginOutList(userId)
					}

					err := tsCon.conn.SetWriteDeadline(time.Now().Add(define.WEBSOCKET_PING_WAIT_TIME * time.Second))
					if err != nil {
						log.Printf("ping error: %s\n", err.Error())
					}
					// 推送消息时加锁，websocket并不支持并发推送
					tsCon.lock.Lock()
					sendMessageData := define.ResponseMessage{
						MessageId:"",
						MessageType:define.MESSAGE_TYPE_PING,
						MessageData:"",
					}
					responseData, _ := json.Marshal(sendMessageData)
					_, err = tsCon.conn.Write(responseData)
					tsCon.lock.Unlock()
					if err != nil {
						// 删除用户长连接
						log.Println("心跳检测:" + userId + "退出登录")
						// 删除长连接
						RemoveUserTcpConnect(userId)
						// 退出登录队列
						cache.SetUserLoginOutList(userId)
					}
					wg.Done()
				}(c, wg)
			}

			wg.Wait()
		}
	}
}

func CheckConnectTsConn() {
	// 每10s清除没有通过验证的tcp sockeet
	connectTs := tcpConnManager.connectTs
	pingPeriod := define.WEBSOCKET_PING_PERIOD * time.Second
	ticker := time.NewTicker(pingPeriod)
	for {
		select {
		case <-ticker.C:
			currentTime := time.Now().Unix()
			log.Println("每10s清除没有通过验证的tcp socket")
			tcpConnManager.lock.RLock()
			for c, t := range connectTs {
				if t < currentTime {
					delete(tcpConnManager.connectTs, c)
					c.conn.Close()
				}
			}
			tcpConnManager.lock.RUnlock()
			break
		}
	}
}

func RemoveUserTcpConnect(userId string) error {
	ts := tcpConnManager.ts
	tcpConn := ts[userId]
	tcpConn.lock.Lock()
	err := tcpConn.conn.Close()
	if err != nil {
		return err
	}
	delete(ts, userId)
	tcpConn.lock.Unlock()
	return nil
}

func ValidTcpConnect(tcpConn *TcpConn, userId string) {
	tcpConnManager.lock.Lock()
	defer tcpConnManager.lock.Unlock()

	tcpConnManager.ts[userId] = tcpConn
	delete(tcpConnManager.connectTs, tcpConn)
}

// Encode 将消息编码
// 前4字节（头部）是消息长度，后面是消息内容（body）
func Encode(message []byte) ([]byte, error) {
	// 读取消息的长度，转换成int32类型（占4个字节）
	var length = int32(len(message))
	var pkg = new(bytes.Buffer)
	// 写入消息头
	err := binary.Write(pkg, binary.LittleEndian, length)
	if err != nil {
		return nil, err
	}
	// 写入消息实体
	// err = binary.Write(pkg, binary.LittleEndian, []byte(message))
	err = binary.Write(pkg, binary.LittleEndian, message)
	if err != nil {
		return nil, err
	}
	return pkg.Bytes(), nil
}

// Decode 解码消息
// 消息定义 {fn:"class.method","data":{"userId":"xxx","message_data":"xxxx"}}
func Decode(reader *bufio.Reader) ([]byte, error) {
	// 读取消息的长度
	lengthByte, _ := reader.Peek(4) // 读取前4个字节的数据
	lengthBuff := bytes.NewBuffer(lengthByte)
	var length int32
	err := binary.Read(lengthBuff, binary.LittleEndian, &length)
	if err != nil {
		return []byte{}, err
	}
	// Buffered返回缓冲中现有的可读取的字节数。
	if int32(reader.Buffered()) < length+4 {
		return []byte{}, err
	}

	// 读取真正的消息数据
	pack := make([]byte, int(4+length))
	_, err = reader.Read(pack)
	if err != nil {
		return []byte{}, err
	}
	// return string(pack[4:]), nil
	return pack[4:], nil
}
