package ANet

import (
	"axj/APro"
	"errors"
	"sync"
)

// CRC错误
var ERR_CLOSED = errors.New("CLOSED")

type Session struct {
	Client    Client
	encryKey  []byte
	closed    bool
	locker    sync.Locker
	handler   Handler
	repQueue  chan RepData
	repBuffer *[]byte
}

type RepData struct {
	req     int32
	uri     string
	data    []byte
	isolate bool
	batch   *RepBatch
}

type Handler interface {
	Processor() Processor
	UriDict() UriDict
	RepQueueSize() int

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
	defer s.Client.Close()
	if s.repQueue != nil {
		defer close(s.repQueue)
	}

	s.closed = true
	if s.handler != nil {
		s.handler.OnClose(s, err, reason)
	}
}

func (s *Session) ReqLoop(handler Handler, poolG APro.PoolG, protocol Protocol, compress Compress, decrypt Encrypt, uriDict UriDict) {
	for {
		if s.closed {
			break
		}

		err, req, uri, uriI, data := Req(s.Client, protocol, compress, decrypt, s.encryKey)
		if err != nil {
			s.Close(err, nil)
			break
		}

		if uri == "" {
			if uriI > 0 && uriDict != nil {
				uri = uriDict.UriIMapUri()[uriI]
			}
		}

		if !handler.OnReq(s, req, uri, data) {
			go s.handlerReqIo(handler, poolG, req, uri, data)
		}
	}
}

func (s *Session) handlerReqIo(handler Handler, poolG APro.PoolG, req int32, uri string, data []byte) {
	if poolG == nil {
		poolG.Add()
		defer poolG.Done()
	}

	handler.OnReqIO(s, req, uri, data)
}

func (s *Session) Rep(req int32, uri string, data []byte, isolate bool, batch *RepBatch) error {
	if s.closed {
		return ERR_CLOSED
	}

	if s.repQueue == nil {
		s.locker.Lock()
		defer s.locker.Unlock()
		if s.closed {
			return ERR_CLOSED
		}

		if s.repQueue == nil {
			s.repQueue = make(chan RepData, s.handler.RepQueueSize())
			go s.repLoop()
		}
	}

	s.repQueue <- RepData{req: req, uri: uri, data: data, isolate: isolate, batch: batch}
	return nil
}

func (s *Session) repLoop() {
	processor := s.handler.Processor()
	uriDict := s.handler.UriDict()
	for {
		data := <-s.repQueue
		if s.closed {
			break
		}

		if data.batch != nil {
			// 批量写入
			err := data.batch.Rep(s.Client, s.repBuffer, processor.Encrypt, s.encryKey)
			if err != nil {
				s.Close(err, nil)
				break
			}

			continue
		}

		uri := data.uri
		var uriI int32 = 0
		if uri != "" && uriDict != nil {
			uriI = uriDict.UriMapUriI()[uri]
			if uriI > 0 {
				uri = ""
			}
		}

		// 单个写入
		err := Rep(s.Client, s.repBuffer, processor.Protocol, processor.Compress, processor.CompressMin, processor.Encrypt, s.encryKey, data.req, uri, uriI, data.data, data.isolate)
		if err != nil {
			s.Close(err, nil)
			break
		}
	}
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
