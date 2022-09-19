package define

const (
	REDIS_SET_USER_ROOM_ID_LIST = "USER:ROOM_ID_" // 用户会话房间列表集合

	REDIS_STRING_USER_WEBSOCKET_CONNECT = "USER:WEBSOCKET_CONNECT_" // 用户接入层路由

	REDIS_LIST_USER_LOGIN_OUT = "USER:LOGIN_OUT" // 用户退出登录消息队列

	REDIS_LIST_USER_SEND_MESSAGE = "USER:SEND_MESSAGE" // 用户发送消息消息队列
)
