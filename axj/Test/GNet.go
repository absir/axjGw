package main

import (
	"axj/ANet"
	"axj/Thrd/Util"
	"net"
	"sync"
	"unsafe"
)

//type example struct{}
//
//func (s *example) OnMessage(c *connection.Connection, ctx interface{}, data []byte) interface{} {
//	return data
//}
//
//func (s *example) OnConnect(c *connection.Connection) {
//	//log.Println(" OnConnect ： ", c.PeerAddr())
//}
//
////func (s *example) OnMessage(c *connection.Connection, ctx interface{}, data []byte) (out []byte) {
////	//log.Println("OnMessage")
////	out = data
////	return
////}
//
//func (s *example) OnClose(c *connection.Connection) {
//	//log.Println("OnClose")
//}

type Test struct {
	wBuff []byte
	idx   int
}

func main() {
	//println(nil)
	pool := new(sync.Pool)
	pool.New = func() interface{} {
		return new(Test)
	}
	for i := 0; i < 3; i++ {
		test := new(Test)
		test.idx = i
		println(unsafe.Pointer(test))
		for j := 0; j < 3; j++ {
			pool.Put(test)
		}
	}

	list := new(Util.ArrayList)
	for i := 0; i < 3; i++ {
		for j := 0; j <= 3; j++ {
			test := pool.Get().(*Test)
			println(test.idx)
			println(unsafe.Pointer(test))
			list.Add(test)
		}
	}

	test := new(Test)
	t := &test.wBuff
	println(t)
	println(t == nil)
	conn := ANet.NewConnSocket(&net.TCPConn{})
	conn.ReadByte()

	//handler := new(example)
	//var port int
	//var loops int
	//
	//flag.IntVar(&port, "port", 1833, "server port")
	//flag.IntVar(&loops, "loops", -1, "num loops")
	//flag.Parse()
	//
	//s, err := gev.NewServer(handler,
	//
	//	gev.Address(":"+strconv.Itoa(port)),
	//	gev.NumLoops(loops))
	//if err != nil {
	//	panic(err)
	//}
	//
	//s.Start()
}
