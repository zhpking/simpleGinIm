package helper

import (
	"errors"
	"go/ast"
	"log"
	"reflect"
	"strings"
	"sync"
)

/*
UserId string `json:"user_id"`
	ToUserId string `json:"to_user_id"`
	RoomId string `json:"room_id"`
	MessageType int64 `json:"message_type"`
	MessageData string `json:"message_data"`
	Fn string `json:"fn"` // 消息处理方法，如service.login
*/
/*
// todo 先简单点枚举出来，实现逻辑后再封装
func AnalyseMessageFn(rm *define.ReceiveMessage, v ...interface{}) {
	switch rm.Fn {
	case "service.userLoginOut":
		user := &service.User{
			UserId:rm.UserId,
			CurrentTime:time.Now().Unix(),
		}
		user.UserLoginOut()
		break
	case "service.kickOutRoom":
		room := &service.Room{
			RoomId:rm.RoomId,
			CurrentTime:time.Now().Unix(),
		}
		room.KickOutRoom()
		break
	case "service.exitRoom":
		room := &service.Room{}
		room.ExitRoom()
		break
	case "service.inviteUserEnterRoom":
		room := &service.Room{}
		room.InviteUserEnterRoom()
		break
	case "service.enterRoom":
		room := &service.Room{}
		room.EnterRoom()
		break
	case "service.createRoom":
		room := &service.Room{}
		room.CreateRoom()
		break
	case "service.singleMessage":
		message := &service.ReceiveMessage{}
		message.SingleMessage()
		break
	case "service.roomMessage":
		message := &service.ReceiveMessage{}
		message.RoomMessage()
		break
	case "service.userSendMessage":
		message := &service.ReceiveMessage{}
		message.UserSendMessage()
		break
	default:
		ret = nil
	}
}
*/

// func (t *T) MethodName(argType T1, replyType *T2) error
type MethodType struct {
	Method reflect.Method // 方法自身对象，即MethodName(reflect.TypeOf(&n).Method())
	ArgType reflect.Type // 第一个参数类型，即argType T1(reflect.TypeOf(&n))
	ReplyType reflect.Type // 第二个参数类型，即replyType *T2(reflect.TypeOf(&n))
}

func (m *MethodType) NewArgv() reflect.Value {
	var argv reflect.Value
	if m.ArgType.Kind() == reflect.Ptr {
		// 如果argType T1是引用类型
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}

	return argv
}

func (m *MethodType) NewReplyv() reflect.Value {
	// replyType *T2 规定是一个引用类型，所以就不需要判断类型了
	replyv := reflect.New(m.ReplyType.Elem())
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}

	return replyv
}

// service 的定义也是非常简洁的，name 即映射的结构体的名称，比如 T，比如 WaitGroup；typ 是结构体的类型；rcvr 即结构体的实例本身，保留 rcvr 是因为在调用时需要 rcvr 作为第 0 个参数；method 是 map 类型，存储映射的结构体的所有符合条件的方法。
type Service struct {
	Name string // 结构体名，即T的字符串
	Typ reflect.Type // 结构体类型，即T反射Type(reflect.TypeOf(&n))
	Rcvr reflect.Value // 结构体的实例本身，即T反射value
	Method map[string]*MethodType
}

// 接下来，完成构造函数 newService，入参是任意需要映射为服务的结构体实例。
func NewService(rcvr interface{}) *Service {
	s := new(Service)
	s.Rcvr = reflect.ValueOf(rcvr)
	s.Name = reflect.Indirect(s.Rcvr).Type().Name()
	s.Typ = reflect.TypeOf(rcvr)
	// 判断结构体名首字母是否大写
	if !ast.IsExported(s.Name) {
		log.Fatalf("rpc server: %s is not a valid service name", s.Name)
	}
	s.RegisterMethods()
	return s
}
// registerMethods 过滤出了符合条件的方法：
//
//两个导出或内置类型的入参（反射时为 3 个，第 0 个是自身，类似于 python 的 self，java 中的 this）
//返回值有且只有 1 个，类型为 error
func (s *Service) RegisterMethods() {
	s.Method = make(map[string]*MethodType)
	for i := 0; i < s.Typ.NumMethod(); i++ {
		method := s.Typ.Method(i)
		mType := method.Type
		// 检查入参和返回值个数
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		// 检查返回值类型
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		// 获取入参两个参数
		argType, replyType := mType.In(1), mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}
		s.Method[method.Name] = &MethodType{
			Method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("rpc server: register %s.%s\n", s.Name, method.Name)
	}
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

// 最后，我们还需要实现 call 方法，即能够通过反射值调用方法
func (s *Service) Call(m *MethodType, argv, replyv reflect.Value) error {
	// atomic.AddUint64(&m.numCalls, 1)
	f := m.Method.Func
	returnValues := f.Call([]reflect.Value{s.Rcvr, argv, replyv})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}



type Server struct {
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func (server *Server) Register(rcvr interface{}) error {
	s := NewService(rcvr)
	if _, dup := server.serviceMap.LoadOrStore(s.Name, s); dup {
		return errors.New("rpc: service already defined:" + s.Name)
	}
	return nil
}

func (server *Server) GetServiceMap() sync.Map {
	return server.serviceMap
}

func (server *Server) FindService(serviceMethod string) (svc *Service, mtype *MethodType, err error) {
	// 获取.的下标
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server: service/method request ill-formed: " + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dot],serviceMethod[dot+1:]
	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can't find service " + serviceName)
		return
	}

	// 接口断言
	svc = svci.(*Service)
	mtype = svc.Method[methodName]
	if mtype == nil {
		err = errors.New("rpc server: can't find method " + methodName)
	}
	return
}

func Register(rcvr interface{}) error {
	return DefaultServer.Register(rcvr)
}

func ServerCall(method string, req interface{}, resp interface{}) error {
	svc, mtype, err := DefaultServer.FindService(method)
	if err != nil {
		log.Printf("[Server Call] %v\n", err.Error())
		return err
	}
	// s := helper.NewService(&foo)
	// mType := s.Method["Sum"]

	argv := mtype.NewArgv()
	replyv := mtype.NewReplyv()
	argv.Set(reflect.ValueOf(req))
	argv.Set(reflect.ValueOf(resp))
	err = svc.Call(mtype, argv, replyv)
	if err != nil {
		log.Printf("[Server Call] %v\n" + err.Error())
		return err
	}

	return nil
}