package ANet

import (
	"axj/APro"
	"axj/Thrd/AZap"
	"errors"
	"go.uber.org/zap"
	"sync"
)

// 错误
var ERR_CLOSED = errors.New("CLOSED")
var ERR_CRASH = errors.New("CRASH")

type Session struct {
	Client    Client
	encryKey  []byte
	closed    bool
	locker    sync.Locker
	handler   Handler
	repBuffer *[]byte
}

type Handler interface {
	Processor() Processor
	UriDict() UriDict
	PoolG() APro.PoolG
	OnReq(session *Session, req int32, uri string, data []byte) bool
	OnReqIO(session *Session, req int32, uri string, data []byte)
	OnClose(session *Session, err error, reason interface{})
}

func (s *Session) Close(err error, reason interface{}) {
	if s.closed {
		return
	}

	s.locker.Lock()
	defer s.locker.Unlock()
	if s.closed {
		return
	}

	// 关闭执行
	defer s.Recover()
	defer s.Client.Close()
	s.closed = true
	// 关闭日志
	AZap.Logger.Info("Recover crash", zap.Error(err), zap.Reflect("reason", reason))
	if s.handler != nil {
		s.handler.OnClose(s, err, reason)
	}
}

func (s *Session) Recover() error {
	if reason := recover(); reason != nil {
		err, ok := reason.(error)
		if ok {
			reason = nil
			s.Close(err, nil)

		} else {
			err = ERR_CRASH
			s.Close(err, reason)
		}

		if reason == nil {
			AZap.Logger.Warn("Recover crash", zap.Error(err))

		} else {
			AZap.Logger.Warn("Recover crash", zap.Error(err), zap.Reflect("reason", reason))
		}

		return err
	}

	return nil
}

func (s *Session) Req() (err error, req int32, uri string, data []byte) {
	if s.closed {
		return ERR_CLOSED, 0, "", nil
	}

	processor := s.handler.Processor()
	var uriI int32 = 0
	err, req, uri, uriI, data = Req(s.Client, processor.Protocol, processor.Compress, processor.Encrypt, s.encryKey)
	if uri == "" && uriI > 0 {
		uriDict := s.handler.UriDict()
		if uriDict != nil {
			uri = uriDict.UriIMapUri()[uriI]
		}
	}

	return nil, req, uri, data
}

func (s *Session) ReqLoop() {
	for {
		if s.closed {
			break
		}

		err, req, uri, data := s.Req()
		if err != nil {
			s.Close(err, nil)
			break
		}

		if !s.handler.OnReq(s, req, uri, data) {
			poolG := s.handler.PoolG()
			if poolG == APro.PoolOne {
				s.handlerReqIo(nil, req, uri, data)

			} else {
				go s.handlerReqIo(poolG, req, uri, data)
				if poolG != nil {
					poolG.Add()
				}
			}
		}
	}
}

func (s *Session) handlerReqIo(poolG APro.PoolG, req int32, uri string, data []byte) {
	if poolG == nil {
		defer poolG.Done()
	}

	s.handler.OnReqIO(s, req, uri, data)
}

// 单进程阻塞 req < 0 直接 WriteData
func (s *Session) Rep(req int32, uri string, data []byte, isolate bool, encry bool, batch *RepBatch) error {
	if s.closed {
		return ERR_CLOSED
	}

	uriDict := s.handler.UriDict()
	var uriI int32 = 0
	if uri != "" && uriDict != nil {
		uriI = uriDict.UriMapUriI()[uri]
		if uriI > 0 {
			uri = ""
		}
	}

	encryKey := s.encryKey
	if !encry {
		encryKey = nil
	}

	processor := s.handler.Processor()
	var err error = nil
	if batch == nil {
		err = Rep(s.Client, s.repBuffer, processor.Protocol, processor.Compress, processor.CompressMin, processor.Encrypt, encryKey, req, uri, uriI, data, isolate)

	} else {
		err = batch.Rep(s.Client, s.repBuffer, processor.Encrypt, encryKey)
	}

	if err != nil {
		s.Close(err, nil)
		return err
	}

	return err
}

func HandlerRepBatch(handler Handler, batch *RepBatch, req int32, uri string, data []byte) *RepBatch {
	processor := handler.Processor()
	uriDict := handler.UriDict()
	var uriI int32 = 0
	if uri != "" && uriDict != nil {
		uriI = uriDict.UriMapUriI()[uri]
		if uriI > 0 {
			uri = ""
		}
	}

	batch.Init(processor.Protocol, processor.Compress, processor.CompressMin, req, uri, uriI, data)
	return batch
}
