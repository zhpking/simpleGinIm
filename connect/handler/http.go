package handler

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
)

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
	// port, _ := helper.GetWsPort()
	// address = address + ":" + port
	// 关联用户登录服务器ip
	cache.SetConnection2User(ut.UserId, address)
	// 登录消息投递
	cache.SetUserLoginList(ut.UserId)

	/*
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
	*/
	// helper.SucResponse(ctx, "系统异常", make(map[string]interface{}))
}

/*
func LoginOut(ctx *gin.Context) {
	// ut := ctx.MustGet("user_token").(*helper.UserToken)
	// RemoveUserConnect(ut.UserId)
	user_id := ctx.PostForm("user_id")
	err := RemoveUserConnect(user_id)
	if err != nil {
		helper.FailResponse(ctx, err.Error())
	} else {
		helper.SucResponse(ctx, "suc", make(map[string]string))
	}
}
*/

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
		helper.FailResponse(ctx, "发送失败")
		return
	}

	// 构造消息体
	sendMessageData := define.ResponseMessage{
		MessageId:messageId,
		MessageType:messageType,
		MessageData:messageData,
	}
	responseData, _ := json.Marshal(sendMessageData)

	wg := &sync.WaitGroup{}
	wg.Add(len(toUserIdList))
	log.Println("一共发送", toUserIdList)
	for _, uId := range toUserIdList {
		// 推送消息(websocket)
		wcCon, ok := wc[uId]
		if ok {
			// 断开连接消息
			if messageType == define.MESSAGE_TYPE_DISCONNECT {
				RemoveUserConnect(uId)
				wg.Done()
				continue
			}

			// err = conn.WriteMessage(websocket.TextMessage, msg.MessageData)
			go func(wcCon *WcConn, wg *sync.WaitGroup, uId string) {
				// 推送消息时加锁，websocket并不支持并发推送
				wcCon.lock.Lock()
				err = wcCon.conn.WriteMessage(websocket.TextMessage, responseData)
				wcCon.lock.Unlock()
				if err != nil {
					log.Printf("%v消息发送失败,error: %v\n", uId, err)
					wg.Done()
					return
				}
				log.Println(uId + "发送成功(websocket)")
				wg.Done()
			}(wcCon, wg, uId)
		}

		// 推送消息 tcp
		tcpConn, ok := tcpConnManager.ts[uId]
		if ok {
			// 断开连接消息
			if messageType == define.MESSAGE_TYPE_DISCONNECT {
				RemoveUserTcpConnect(uId)
				wg.Done()
				continue
			}

			go func(tcpConn *TcpConn, wg *sync.WaitGroup, uId string) {
				// 推送消息时加锁，websocket并不支持并发推送
				tcpConn.lock.Lock()
				msg, err := Encode(responseData)
				if err != nil {
					log.Printf("[TCP ENCODE ERROR] %v err: %v\n", uId, msg)
				}
				_, err = tcpConn.conn.Write(msg)
				tcpConn.lock.Unlock()
				if err != nil {
					log.Printf("%v消息发送失败,error: %v\n", uId, err)
					// helper.FailResponse(ctx, "发送失败")
					helper.FailResponse(ctx, "发送失败")
					wg.Done()
					return
				}
				log.Println(uId + "发送成功(tcp)")
				wg.Done()
			}(tcpConn, wg, uId)
		}
	}

	wg.Wait()
	helper.SucResponse(ctx, "发送成功", make(map[string]string))
}