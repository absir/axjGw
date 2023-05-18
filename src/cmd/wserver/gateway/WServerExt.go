package gateway

import (
	"axj/ANet"
	"axj/Kt/KtBuffer"
	"axjGW/pkg/gateway"
	"encoding/json"
	"golang.org/x/net/websocket"
)

type ProcessorExt struct {
	ANet.ProcessorV
}

func (that *ProcessorExt) ReqOpen(i int, pBuffer **KtBuffer.Buffer, conn ANet.Conn, decryKey []byte) (error, int32, string, int32, []byte) {
	if web, ok := conn.(*ANet.ConnWebsocket); ok {
		if i == 0 {
			return nil, 0, "", 0, nil

		} else if i == 1 {
			req := web.Conn().Request()
			data, err := json.Marshal([]interface{}{req.RequestURI, req.Header})
			// Kt.Log(KtUnsafe.BytesToString(data))
			return err, 0, "", 0, data
		}
	}

	return that.ProcessorV.ReqOpen(i, pBuffer, conn, decryKey)
}

func (that *ProcessorExt) Req(pBuffer **KtBuffer.Buffer, conn ANet.Conn, decryKey []byte) (error, int32, string, int32, []byte) {
	if _, ok := conn.(*ANet.ConnWebsocket); ok {
		err, bs, _ := conn.ReadA()
		return err, ANet.REQ_ONEWAY, "S", 0, bs
	}

	return that.ProcessorV.Req(pBuffer, conn, decryKey)
}

func (that *ProcessorExt) Rep(bufferP bool, conn ANet.Conn, encryKey []byte, compress bool, req int32, uri string, uriI int32, data []byte, isolate bool, id int64) error {
	if web, ok := conn.(*ANet.ConnWebsocket); ok {
		if req < 0 {
			// websocket心跳
			if _, ok := conn.(*ANet.ConnWebsocket); ok {
				return nil
			}

		} else if req > 0 {
			return nil
		}

		if uri != "" && req != ANet.REQ_LOOP {
			return websocket.Message.Send(web.Conn(), uri)
		}

		if data != nil {
			return conn.Write(data)
		}

		return nil
	}

	return that.ProcessorV.Rep(bufferP, conn, encryKey, compress, req, uri, uriI, data, isolate, id)
}

func InitGateWay() {
	processor := &ProcessorExt{
		ProcessorV: ANet.ProcessorV{
			Protocol:    &ANet.ProtocolV{},
			CompressMin: gateway.Config.CompressMin,
			DataMax:     gateway.Config.DataMax,
		},
	}

	gateway.Processor = processor
}
