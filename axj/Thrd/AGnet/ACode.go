package AGnet

import (
	"axj/ANet"
	"github.com/panjf2000/gnet"
)

type ACode struct {
	processor *ANet.Processor
}

func NewACode(processor *ANet.Processor) *ACode {
	that := new(ACode)
	that.processor = processor
	return that
}

func (that ACode) Encode(c gnet.Conn, buf []byte) ([]byte, error) {
	return buf, nil
}

func (that ACode) Decode(c gnet.Conn) ([]byte, error) {
	conn := connCtx(c)
	if conn != nil {
		bs := c.Read()
		c.ResetBuffer()
		pBs := &bs
		for {
			err := that.processor.Protocol.ReqFrame(pBs, conn.frameReader, that.processor.DataMax)
			if err != nil {
				return nil, err
			}

			frame := conn.frameReader.DoneFrame()
			if frame == nil {
				break
			}

			// 加入缓冲区
			conn.locker.Lock()
			if conn.closed {
				// 已关闭
				conn.locker.Unlock()
				return nil, ANet.ERR_CLOSED
			}

			conn.listAsync.SubmitLock(frame, false)
			conn.locker.Unlock()
			// frame最大缓冲长度
			if conn.listAsync.Size() > ANet.FRAME_MAX {
				return nil, ANet.ERR_FRAME_MAX
			}
		}
	}

	return nil, nil
}
