# simpleGinIm

基于gin，一个简单的，给第三方项目接入的IM系统，目前该项目在逐步完善中，具体请看更新日志

## Go版本

\>= go 1.17

## 关于系统

业务层监听端口为8080，如果需要可以自行进行修改

接入层监听端口为18080，如果需要可以自行进行修改

## 系统目录说明

```

.
├── api                     # 业务层接口
├── cache                   # 缓存方法定义
├── config                  # 配置文件
├── connect                 # 接入层接口
├── define                  # 常量层，用于定义一些枚举或者通用的数据结构
├── example                 # 一些实例代码
├── helper                  # 封装的一些常用函数
├── middleware              # gin中间件
├── model                   # 数据库模型
├── service                 # 通用逻辑层
├── test                    # 单元测试
├── go.mod                  # go包管理mod文件
├── go.sum                  # go包管理sum文件
├── main.go                 # 程序唯一入口文件
└── readme.md               # 说明文档

```

## 更新日志

- 2022-10-04

    - 支持socket接入方式
    - 添加心跳检查
    - 接入层提供推送消息接口，支持业务层进行消息推送
    - 使用redis作为消息队列，解耦业务层和接入层
    - 优化项目目录结构

- 2022-09-19

    - 接入层和业务层拆分，两者间用http请求和消息队列进行通讯
    - 优化了群聊和单聊，群聊使用的是广播的方式，广播所有的接入层推送消息
    - 目前系统还只是支持websocket接入（web），下一步计划加上socket接入（android和ios）

- 2022-09-11

    - 上传第一版本，支持私聊和房间广播功能
    - 支持websocket接入方式，并提供http接口对外提供服务
    - 目前系统还是个大单体，下一步计划拆分接入层和业务层

## 数据库

项目目前用到的是mongodb和redis，在使用本系统之前，请先安装好

### mongo 集合说明

1. message 消息集合

```json

{
  "message_id": "消息id",
  "message_data": "消息数据",
  "created_at": "创建时间",
  "message_status": "消息状态【1正常】"
}

```

2. user 用户集合

```json

{
  "user_id": "系统接入的用户id",
  "login_status": "登录状态【0离线1在线】",
  "last_login_time": "最后一次登录时间",
  "last_login_out_time": "最后一个退出时间",
  "user_status": "用户状态【1正常-1删除】",
  "created_at": "创建时间"
}

```

3. room 房间（会话）集合

```json

{
  "room_id": "房间id",
  "room_type": "房间类型【1单人2多人】",
  "room_status": "房间状态【1正常-1删除】",
  "last_user_id": "最后一个发消息的用户id",
  "last_message_id": "最后一条消息id",
  "last_message_updated_at": "最后一条消息更新时间",
  "created_at": "创建时间"
}

```

4. user_room 用户所在房间关联集合

```json

{
  "user_id": "用户id",
  "room_id": "房间id",
  "created_at": "创建时间"
}

```

5. user_message 用户消息关联集合

```json

{
  "room_id": "房间id",
  "message_id": "消息id",
  "send_user_id": "发送人id",
  "created_at": "创建时间",
  "send_status": "发送状态【1发送成功-1撤回】"
}

```

6. user_room_chat_log 用户参与过的房间聊天关联集合

```json

{
  "user_id": "用户id",
  "room_id": "房间id",
  "created_at": "创建时间",
  "updated_at": "更新时间"
}

```

## 配置文件说明

配置文件都在根目录下的config目录内

```

[mongo]
address = mongodb连接地址
username = 用户名
password = 密码
database = 库名

[redis]
address = redis连接地址

[login]
loginExpire = 登录超时时间，单位为秒，0为不超时

[api]
address=业务层ip地址，多个ip用逗号分隔，如 192.168.78.134,192.168.78.135
port=业务层监听端口

[ws]
address=接入层ip地址，多个ip用逗号分隔，如 192.168.78.134,192.168.78.135
port=接入层监听端口

```

## API

### 接口地址说明

符号				|说明
:----:			|:---
ip				|请求域名
port			|请求端口

### 格式说明

符号				|说明
:----:			|:---
R				|报文中该元素必须出现（Required）
O				|报文中该元素可选出现（Optional）


### 参数说明

名称				|描述			                                                    |备注  
:----			|:---		                                                        |:---	
公共参数			|每个接口都包含的通用参数，以JSON格式存放在Header属性		                |用户登录后token，没有登录则为空字符串
私有参数			|每个接口特有的参数，其中POST方式放在Body属性，GET方式放在Params数据		|用户登录后token，没有登录则为空字符串

#### 公共参数

公共参数（Header）是用于接口鉴权的参数，每次请求均需要携带这些参数：

参数名称				|类型		|出现要求	|描述  
:----				|:---		|:------	|:---	
token				|string		|R			|用户登录后token，没有登录则为空字符串

### 接口定义

#### 获取登录token

##### 接口地址

> [POST]https//ip:port/user/get_login_token 

##### 请求参数

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
Body						|&nbsp;		|R			|&nbsp;
&emsp;user_id				|string		|R			|接入IM的系统用户id

##### 返回结果

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
code						|int		|R			|响应码
msg							|string		|R			|&nbsp;
data						|object		|R			|&nbsp;
&emsp;token				    |string		|R			|用户token
&emsp;tcpAddress			|string		|R			|tcp连接ip
&emsp;tcpPort				|string		|R			|tcp连接端口
&emsp;websocketAddress		|string		|R			|websocket连接ip
&emsp;websocketPort			|string		|R			|websocket连接端口

##### 请求示例

```json

{
    "Header":{
    },
    "Body":{
        "user_id":"18520322032"
    }
}

```

##### 返回结果示例

```json

{
    "code": 200,
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTAwMDAyMDAyNzUiLCJsb2dpbl9leHBpcmUiOjB9.bHWKr0XSOxLwUHd1os-AltV-btgtAVyRCuaGGVoP2io",
        "tcpAddress": "192.168.0.1",
        "tcpPort": "18080",
        "websocketAddress": "192.168.0.1",
        "websocketPort": "28080"
    },
    "msg": "获取token成功"
}

```

##### 当token失效的时候，则返回

```json

{
    "code": -1,
    "msg": "token已失效"
}

```



#### 退出登录

##### 接口地址

> [POST]https//ip:port/u/user/login_out

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
Header						|&nbsp;		|R			|请求报文头
&emsp;token					|string		|R			|用户登录后token，没有登录则为空字符串

##### 返回结果

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
code						|int		|R			|响应码
msg							|string		|R			|&nbsp;
data						|object		|R			|&nbsp;

##### 请求示例

```json

{
    "Header":{
        "token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTAwMDAyMDAwMDQiLCJsb2dpbl9leHBpcmUiOjB9.LPIj_rqyy1gxl_upNfSGtVuhoEmW2GOhc06Wz7GHOSY"
    },
    "Body":{
    }
}

```

##### 返回结果示例

```json

{
    "code": 200,
    "data": {},
    "msg": "退出登录成功"
}

```



#### 连接IM

##### 接口地址

> ws://ip:port/u/user/connect

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
Header						|&nbsp;		|R			|请求报文头
&emsp;token					|string		|R			|用户登录后token，没有登录则为空字符串

##### 发送消息数据格式

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
user_id						|string		|R			|发送者用户id
to_user_id					|string		|R			|接收者用户id				|object		|R			|&nbsp;
room_id				        |string		|R			|房间id
message_type				|int		|R			|消息类型
message_data				|string		|R			|消息内容

```json

{
	"user_id":"10000200275",
	"to_user_id":"10000200117",
	"room_id":"39336433396564352d393138322d343062632d626330382d333134313064616561356234",
	"message_type":1,
	"message_data":"hello"
}

```

##### 接收消息数据格式

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
message_id					|string		|R			|消息id
message_type				|int		|R			|消息类型
message_data				|string		|R			|消息内容

```json

{
	"message_id": "66363761663363642d303830342d346339622d383131332d393864326532376163616132",
	"message_type": 1,
	"message_data": "hello"
}

```

#### 创建群聊房间

##### 接口地址

> [POST]https//ip:port/rm/room/create_room

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
Header						|&nbsp;		|R			|请求报文头
&emsp;token					|string		|R			|用户登录后token，没有登录则为空字符串

##### 返回结果

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
code						|int		|R			|响应码
msg							|string		|R			|&nbsp;
data						|object		|R			|&nbsp;
&emsp;roomId				|string		|R			|房间id

##### 请求示例

```json

{
    "Header":{
            "token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTAwMDAyMDAwMDQiLCJsb2dpbl9leHBpcmUiOjB9.LPIj_rqyy1gxl_upNfSGtVuhoEmW2GOhc06Wz7GHOSY"
    },
    "Body":{
        "user_id":"18520322032"
    }
}

```

##### 返回结果示例

```json

{
    "code": 200,
    "data": {
        "roomId": "39336433396564352d393138322d343062632d626330382d333134313064616561356234"
    },
    "msg": "创建成功"
}

```



#### 用户进入群聊房间

##### 接口地址

> [POST]https//ip:port/rm/room/enter_room

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
Header						|&nbsp;		|R			|请求报文头
&emsp;token					|string		|R			|用户登录后token，没有登录则为空字符串
Body						|&nbsp;		|R			|&nbsp;
&emsp;room_id			    |string		|R			|房间id

##### 返回结果

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
code						|int		|R			|响应码
msg							|string		|R			|&nbsp;
data						|object		|R			|&nbsp;

##### 请求示例

```json

{
    "Header":{
        "token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTAwMDAyMDAwMDQiLCJsb2dpbl9leHBpcmUiOjB9.LPIj_rqyy1gxl_upNfSGtVuhoEmW2GOhc06Wz7GHOSY"
    },
    "Body":{
        "room_id":"39336433396564352d393138322d343062632d626330382d333134313064616561356234"
    }
}

```

##### 返回结果示例

```json

{
    "code": 200,
    "data": {},
    "msg": "进入房间成功"
}

```



#### 邀请用户进入房间

##### 接口地址

> [POST]https//ip:port/rm/room/invite_user_enter_room

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
Header						|&nbsp;		|R			|请求报文头
&emsp;token					|string		|R			|用户登录后token，没有登录则为空字符串
Body						|&nbsp;		|R			|&nbsp;
&emsp;room_id			    |string		|R			|房间id
&emsp;invite_user_id		|string		|R			|邀请用户id

##### 返回结果

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
code						|int		|R			|响应码
msg							|string		|R			|&nbsp;
data						|object		|R			|&nbsp;

##### 请求示例

```json

{
    "Header":{
        "token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTAwMDAyMDAwMDQiLCJsb2dpbl9leHBpcmUiOjB9.LPIj_rqyy1gxl_upNfSGtVuhoEmW2GOhc06Wz7GHOSY"
    },
    "Body":{
        "invite_user_id":"18231234353",
        "room_id":"39336433396564352d393138322d343062632d626330382d333134313064616561356234"
    }
}

```

##### 返回结果示例

```json

{
    "code": 200,
    "data": {},
    "msg": "邀请用户进入房间成功"
}

```



#### 用户退出群聊房间

##### 接口地址

> [POST]https//ip:port/rm/room/exit_room

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
Header						|&nbsp;		|R			|请求报文头
&emsp;token					|string		|R			|用户登录后token，没有登录则为空字符串
Body						|&nbsp;		|R			|&nbsp;
&emsp;room_id			    |string		|R			|房间id

##### 返回结果

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
code						|int		|R			|响应码
msg							|string		|R			|&nbsp;
data						|object		|R			|&nbsp;

##### 请求示例

```json

{
    "Header":{
        "token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTAwMDAyMDAwMDQiLCJsb2dpbl9leHBpcmUiOjB9.LPIj_rqyy1gxl_upNfSGtVuhoEmW2GOhc06Wz7GHOSY"
    },
    "Body":{
        "room_id":"39336433396564352d393138322d343062632d626330382d333134313064616561356234"
    }
}

```

##### 返回结果示例

```json

{
    "code": 200,
    "data": {},
    "msg": "退出房间成功"
}

```



#### 将用户移出房间

##### 接口地址

> [POST]https//ip:port/rm/room/kick_out_room

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
Header						|&nbsp;		|R			|请求报文头
&emsp;token					|string		|R			|用户登录后token，没有登录则为空字符串
Body						|&nbsp;		|R			|&nbsp;
&emsp;room_id			    |string		|R			|房间id
&emsp;user_id			    |string		|R			|移出房间用户id

##### 返回结果

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
code						|int		|R			|响应码
msg							|string		|R			|&nbsp;
data						|object		|R			|&nbsp;

##### 请求示例

```json

{
    "Header":{
        "token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTAwMDAyMDAwMDQiLCJsb2dpbl9leHBpcmUiOjB9.LPIj_rqyy1gxl_upNfSGtVuhoEmW2GOhc06Wz7GHOSY"
    },
    "Body":{
        "user_id":"18534275932",
        "room_id":"39336433396564352d393138322d343062632d626330382d333134313064616561356234"
    }
}

```

##### 返回结果示例

```json

{
    "code": 200,
    "data": {},
    "msg": "踢出用户成功"
}

```


#### 获取会话列表

##### 接口地址

> [GET]https//ip:port/rm/room/get_room_list

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
Header						|&nbsp;		|R			|请求报文头
&emsp;token					|string		|R			|用户登录后token，没有登录则为空字符串
Params						|&nbsp;		|R			|&nbsp;
&emsp;room_type			    |int		|O			|房间类型
&emsp;page			        |int		|O			|第x页
&emsp;page_size			    |int		|O			|每页获取的条数

##### 返回结果

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
code						|int		|R			|响应码
msg							|string		|R			|&nbsp;
data						|object		|R			|&nbsp;
&emsp;RoomId			    |string		|R			|房间id
&emsp;RoomType			    |int		|R			|房间类型
&emsp;RoomStatus			|int		|R			|房间状态
&emsp;RoomHostUserId		|string		|R			|房主用户id
&emsp;LastUserId			|string		|R			|最后一次发言用户id
&emsp;LastMessageId			|string		|R			|最后一条消息id
&emsp;LastMessageUpdatedAt	|int		|R			|最后一条消息发送时间
&emsp;CreateAt			    |int		|R			|房间创建时间


##### 请求示例

```json

{
    "Header":{
        "token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTAwMDAyMDAwMDQiLCJsb2dpbl9leHBpcmUiOjB9.LPIj_rqyy1gxl_upNfSGtVuhoEmW2GOhc06Wz7GHOSY"
    },
    "Body":{
    }
}

```

##### 返回结果示例

```json

{
    "code": 200,
    "data": [
        {
            "RoomId": "31363966353833312d666130362d343961372d396436652d653236353030643963656436",
            "RoomType": 1,
            "RoomStatus": 1,
            "RoomHostUserId": "10000200275",
            "LastUserId": "10000200275",
            "LastMessageId": "66363761663363642d303830342d346339622d383131332d393864326532376163616132",
            "LastMessageUpdatedAt": 1662886519,
            "CreateAt": 1662885163
        }
    ],
    "msg": "获取成功"
}

```



#### 获取某个会话聊天记录

##### 接口地址

> [GET] https//ip:port/rm/room/get_room_message_list

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
Header						|&nbsp;		|R			|请求报文头
&emsp;token					|string		|R			|用户登录后token，没有登录则为空字符串
Params						|&nbsp;		|R			|&nbsp;
&emsp;room_id			    |string		|R			|房间id
&emsp;page			        |int		|O			|第x页
&emsp;page_size			    |int		|O			|每页获取的条数

##### 返回结果

参数名称						|类型		|出现要求	|描述  
:----						|:---		|:------	|:---	
code						|int		|R			|响应码
msg							|string		|R			|&nbsp;
data						|object		|R			|&nbsp;
&emsp;RoomId			    |string		|R			|房间id
&emsp;MessageId			    |string		|R			|消息id
&emsp;SendUserId			|int		|R			|消息发送用户id
&emsp;SendStatus		    |int		|R			|发送状态
&emsp;CreateAt			    |int		|R			|消息发送时间


##### 请求示例

```json

{
    "Header":{
        "token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTAwMDAyMDAwMDQiLCJsb2dpbl9leHBpcmUiOjB9.LPIj_rqyy1gxl_upNfSGtVuhoEmW2GOhc06Wz7GHOSY"
    },
    "Body":{
    }
}

```

##### 返回结果示例

```json

{
    "code": 200,
    "data": [
        {
            "RoomId": "31363966353833312d666130362d343961372d396436652d653236353030643963656436",
            "MessageId": "38323965356235662d376337312d343735652d393131312d383032653463653363656337",
            "SendUserId": "10000200275",
            "SendStatus": 1,
            "CreateAt": 1662893077
        },
        {
            "RoomId": "31363966353833312d666130362d343961372d396436652d653236353030643963656436",
            "MessageId": "34653331396366632d656162382d343437362d396563632d393737636365323136616132",
            "SendUserId": "10000200275",
            "SendStatus": 1,
            "CreateAt": 1662893077
        },
        {
            "RoomId": "31363966353833312d666130362d343961372d396436652d653236353030643963656436",
            "MessageId": "38663464376530332d613363612d343031322d616664372d326630613732343566386130",
            "SendUserId": "10000200275",
            "SendStatus": 1,
            "CreateAt": 1662893075
        },
        {
            "RoomId": "31363966353833312d666130362d343961372d396436652d653236353030643963656436",
            "MessageId": "65643033306633662d623034382d343336372d396633322d363831643166356463333465",
            "SendUserId": "10000200275",
            "SendStatus": 1,
            "CreateAt": 1662893002
        }
    ],
    "msg": "获取成功"
}

```
