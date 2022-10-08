package handler

import (
	"encoding/json"
	"log"
	"github.com/gorilla/websocket"
	"simpleGinIm/cache"
	"simpleGinIm/define"
	"sync"
	"time"
)

var upgrader websocket.Upgrader = websocket.Upgrader{}
// var wc map[string]*websocket.Conn = make(map[string]*websocket.Conn)
var wc map[string]*WcConn = make(map[string]*WcConn)

type WcConn struct {
	conn *websocket.Conn
	expireTime int64
	lock sync.Mutex
}

func RemoveUserConnect(userId string) error {
	wcConn := wc[userId]
	wcConn.lock.Lock()
	err := wcConn.conn.Close()
	if err != nil {
		return err
	}
	delete(wc, userId)
	wcConn.lock.Unlock()
	return nil
}

// websocket心跳参考 https://www.cnblogs.com/tianyun5115/p/12613274.html
func CheckWebSocketConn() {
	pingPeriod := define.WEBSOCKET_PING_PERIOD * time.Second
	ticker := time.NewTicker(pingPeriod)

	for {
		select {
		case <-ticker.C:
			log.Println("开始心跳检测")
			currentTime := time.Now().Unix()
			if len(wc) == 0 {
				return
			}

			wg := &sync.WaitGroup{}
			wg.Add(len(wc))
			for userId, c := range wc {
				// 超时登录&心跳检测
				go func(wcCon *WcConn, wg *sync.WaitGroup) {
					if wcCon.expireTime != 0 && wcCon.expireTime < currentTime {
						// todo 登录超时，退出登录
						log.Println("心跳检测:" + userId + "退出登录")
						// 删除用户连接
						RemoveUserConnect(userId)
						// 需要退出登录的用户id放进消息队列
						cache.SetUserLoginOutList(userId)
					}

					err := wcCon.conn.SetWriteDeadline(time.Now().Add(define.WEBSOCKET_PING_WAIT_TIME * time.Second))
					if err != nil {
						log.Printf("ping error: %s\n", err.Error())
					}
					// 推送消息时加锁，websocket并不支持并发推送
					wcCon.lock.Lock()
					// err = wcCon.conn.WriteMessage(websocket.PingMessage, nil)
					// 构造消息体
					sendMessageData := define.ResponseMessage{
						MessageId:"",
						MessageType:define.MESSAGE_TYPE_PING,
						MessageData:"",
					}
					responseData, _ := json.Marshal(sendMessageData)
					// 发送ping消息
					err = wcCon.conn.WriteMessage(websocket.TextMessage, responseData)
					wcCon.lock.Unlock()
					if err != nil {
						// 登录超时，退出登录
						log.Println("心跳检测:" + userId + "退出登录")
						// 删除用户连接
						RemoveUserConnect(userId)
						// 需要退出登录的用户id放进消息队列
						cache.SetUserLoginOutList(userId)
					}
					wg.Done()
				}(c, wg)
			}

			wg.Wait()
		}
	}
}
