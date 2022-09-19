package define

const (
	MESSAGE_STATUS_OK = 1 // 消息状态：正常
	MESSAGE_STATUS_CANCEL = 0 // 消息状态：撤回

	MESSAGE_TYPE_TEXT = 1; // 消息类型：文本消息
)

// 单聊推送消息体
type ResponseMessage struct {
	MessageId string `json:"message_id"`
	MessageType int64 `json:"message_type"`
	MessageData string `json:"message_data"`
}

// 群聊推送消息体
type ManyResponseMessage struct {
	ResponseMessage ResponseMessage `json:"response_message"`
	UserIdList []string  `json:"user_id_list"`
}

// 接收消息消息体
type ReceiveMessage struct {
	UserId string `json:"user_id"`
	ToUserId string `json:"to_user_id"`
	RoomId string `json:"room_id"`
	MessageType int64 `json:"message_type"`
	MessageData string `json:"message_data"`
}