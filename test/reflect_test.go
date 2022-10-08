package test

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"simpleGinIm/helper"
	"simpleGinIm/service"
	"strings"
	"testing"
)

func TestReflect(t *testing.T) {
	var n net.OpError
	// 反射对象
	typ := reflect.TypeOf(&n)

	// typ.NumMethod() 获取方法数量
	for i := 0; i < typ.NumMethod(); i ++ {
		// typ.Method(i) 获取对应的方法对象
		method := typ.Method(i)
		// 入参
		argv := make([]string, 0, method.Type.NumIn())
		// 返回值
		returns := make([]string, 0, method.Type.NumOut())
		// j从1开始，第0个入参是n对象本身
		for j := 1; j < method.Type.NumIn(); j ++ {
			argv = append(argv, method.Type.In(j).Name())
		}
		for j := 0; j < method.Type.NumOut(); j ++ {
			returns = append(returns, method.Type.Out(j).Name())
		}

		log.Printf("func (r *%s) %s(%s) %s",
			typ.Elem().Name(),
			method.Name,
			strings.Join(argv, ","),
			strings.Join(returns, ","),
			)
	}
}

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

// it's not a exported Method
func (f Foo) sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func _assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+msg, v...))
	}
}

func TestNewService(t *testing.T) {
	var foo Foo
	s := helper.NewService(&foo)
	_assert(len(s.Method) == 1, "wrong service Method, expect 1, but got %d", len(s.Method))
	mType := s.Method["Sum"]
	_assert(mType != nil, "wrong Method, Sum shouldn't nil")
}

func TestMethodType_Call(t *testing.T) {
	var foo Foo
	s := helper.NewService(&foo)
	mType := s.Method["Sum"]

	argv := mType.NewArgv()
	replyv := mType.NewReplyv()
	argv.Set(reflect.ValueOf(Args{Num1: 1, Num2: 3}))
	err := s.Call(mType, argv, replyv)
	_assert(err == nil && *replyv.Interface().(*int) == 4, "failed to call Foo.Sum")
}

func TestMethodType_Register(t *testing.T) {
	// var foo Foo
	// room := service.Room{RoomId:"roomIdTest"}
	str := "Room.GetRoomReflectTest"
	server := helper.NewServer()
	// _ = helper.NewService(&room)
	var room service.Room = service.Room{RoomId:"testRoomId"}
	err := server.Register(&room)
	if err != nil {
		t.Fatal("11111" + err.Error())
	}

	svc, mtype, err := server.FindService(str)
	if err != nil {
		t.Fatal("22222" + err.Error())
	}
	// s := helper.NewService(&foo)
	// mType := s.Method["Sum"]

	argv := mtype.NewArgv()
	replyv := mtype.NewReplyv()
	argv.Set(reflect.ValueOf(service.RoomParams{UserId:"10000200275"}))
	err = svc.Call(mtype, argv, replyv)
	if err != nil {
		t.Fatal("33333" + err.Error())
	}
	// _assert(err == nil && *replyv.Interface().(*int) == 4, "failed to call Foo.Sum")
}
