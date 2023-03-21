package ANet

import (
	"axj/Thrd/Util"
	"io"
	"sync"
)

type ConnPoll struct {
	conn        Conn
	locker      sync.Locker
	listAsync   *Util.ListAsync
	frameReader *FrameReader
	client      Client
}

func NewConnPoll(conn Conn) *ConnPoll {
	that := new(ConnPoll)
	that.conn = conn
	that.locker = new(sync.Mutex)
	that.listAsync = Util.NewListAsync(nil, that.locker)
	that.frameReader = new(FrameReader)
	return that
}

func (that *ConnPoll) IsClose() bool {
	return that.conn == nil
}

func (that *ConnPoll) OnRead(processor *ProcessorV, bs []byte, bufferP bool) error {
	pBs := &bs
	for {
		err := processor.Protocol.ReqFrame(pBs, that.frameReader, processor.DataMax, bufferP)
		if err != nil {
			return err
		}

		frame := that.frameReader.DoneFrame()
		if frame == nil {
			break
		}

		// 加入缓冲区
		that.locker.Lock()
		if that.IsClose() {
			// 已关闭
			that.locker.Unlock()
			return ERR_CLOSED
		}

		that.listAsync.SubmitLock(frame, false)
		that.locker.Unlock()
		// frame最大缓冲长度
		if that.listAsync.Size() > FRAME_MAX {
			return ERR_FRAME_MAX
		}
	}

	return nil
}

func (that *ConnPoll) FrameStart(fun func(interface{})) {
	that.listAsync.SetRun(fun)
	that.listAsync.Start()
}

func (that *ConnPoll) FrameReq(client Client, frame *ReqFrame) {
	that.client = client
	that.listAsync.SetRun(that.frameReqRun)
	if frame != nil {
		that.frameReqRun(frame)
	}
}

func (that *ConnPoll) frameReqRun(el interface{}) {
	// 请求客户端
	client := that.client
	if client == nil {
		that.OnClose()
		return
	}

	frame, _ := el.(*ReqFrame)
	if frame == nil {
		// EOF close
		client.Get().Close(io.EOF, nil)
		return
	}

	// 请求处理
	client.Get().OnReq(client.Get().ReqFrame(frame))
}

func (that *ConnPoll) OnClose() {
	client := that.client
	if client != nil {
		// client先关闭
		that.client = nil
		client.Get().Close(nil, nil)
	}

	conn := that.conn
	if conn == nil {
		return
	}

	that.locker.Lock()
	conn = that.conn
	if conn == nil {
		that.locker.Unlock()
		return
	}

	that.conn = nil
	that.listAsync.Clear(false)
	that.locker.Unlock()
	conn.Close(true)
}
