package service

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"simpleGinIm/cache"
	"simpleGinIm/define"
	"simpleGinIm/helper"
	"strconv"
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

/*
type responseMessage struct {
	MessageId string `json:"message_id"`
	MessageType int64 `json:"message_type"`
	MessageData string `json:"message_data"`
}
*/

func Connect(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		helper.FailResponse(ctx, "系统异常")
		conn.Close()
		return
	}

	ut := ctx.MustGet("user_token").(*helper.UserToken)
	wcConn := &WcConn{
		conn:conn,
		expireTime:ut.LoginExpire,
	}
	wc[ut.UserId] = wcConn

	// 建立用户socket和服务器ip的路由映射
	ipList := helper.GetLocalIP()
	address := ipList[0]
	// 获取端口
	port, _ := helper.GetWsPort()
	address = address + ":" + port
	// 建立用户id和长连接socket的关系
	cache.SetConnection2User(ut.UserId, address)

	// 接收客户端发送过来的消息
	for {
		// https://www.jianshu.com/p/5d000523e2bd
		receiveMsg := &define.ReceiveMessage{}
		err := conn.ReadJSON(receiveMsg)
		if err != nil {
			log.Printf("[MESSAGE RECEIVE ERROR]  websocket.go, err:接收用户消息错误, err:%v\n", err)
			continue
		}

		// 转为json，扔给消息队列
		data, err := json.Marshal(receiveMsg)
		if err != nil {
			log.Printf("[JSON ERROR] websocket.go, err:%v\n", err)
		}
		cache.SetUserSendMessageList(string(data))
	}
	// helper.SucResponse(ctx, "系统异常", make(map[string]interface{}))
}

func LoginOut(ctx *gin.Context) {
	ut := ctx.MustGet("user_token").(*helper.UserToken)
	RemoveUserConnect(ut.UserId)
}

// 推送消息
func PushSingleMessage(ctx *gin.Context) {
	toUserId := ctx.PostForm("to_user_id")
	messageId := ctx.PostForm("message_id")
	messageTypeStr := ctx.PostForm("message_type")
	messageData := ctx.PostForm("message_data")
	messageType, _ := strconv.ParseInt(messageTypeStr, 10, 64)

	toUserIdList := []string{}
	err := json.Unmarshal([]byte(toUserId), &toUserIdList)
	if err != nil {
		log.Printf("[JSON ERROR] json转化错误 %v\n", err)
		return
	}

	// 构造消息体
	sendMessageData := define.ResponseMessage{
		MessageId:messageId,
		MessageType:messageType,
		MessageData:messageData,
	}
	responseData, _ := json.Marshal(sendMessageData)

	for _, uId := range toUserIdList {
		// 推送消息
		wcCon, ok := wc[uId]
		if !ok {
			// 对方不在线
			// helper.SucResponse(ctx, "发送成功", make(map[string]interface{}))
			continue
		}

		// err = conn.WriteMessage(websocket.TextMessage, msg.MessageData)
		go func(wcCon *WcConn) {
			// 推送消息时加锁，websocket并不支持并发推送
			wcCon.lock.Lock()
			err = wcCon.conn.WriteMessage(websocket.TextMessage, responseData)
			wcCon.lock.Unlock()
			if err != nil {
				log.Printf("消息发送失败,error: %v\n", err)
				helper.FailResponse(ctx, "发送失败")
				return
			}
		}(wcCon)
	}
}

func RemoveUserConnect(userId string) error {
	err := wc[userId].conn.Close()
	if err != nil {
		return err
	}
	delete(wc, userId)
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
						// todo 把需要退出登录的用户id放进消息队列
						RemoveUserConnect(userId)
					}

					err := wcCon.conn.SetWriteDeadline(time.Now().Add(define.WEBSOCKET_PING_WAIT_TIME * time.Second))
					if err != nil {
						log.Printf("ping error: %s\n", err.Error())
					}
					// 推送消息时加锁，websocket并不支持并发推送
					wcCon.lock.Lock()
					err = wcCon.conn.WriteMessage(websocket.PingMessage, nil)
					wcCon.lock.Unlock()
					if err != nil {
						// todo 登录超时，退出登录
						// 删除用户连接
						log.Println("心跳检测:" + userId + "退出登录")
						// todo 把需要退出登录的用户id放进消息队列
						RemoveUserConnect(userId)
					}
					wg.Done()
				}(c, wg)
			}

			wg.Wait()
		}
	}
}
