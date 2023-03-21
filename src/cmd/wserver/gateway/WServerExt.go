package gateway

import (
	"axj/ANet"
	"axj/Kt/KtBuffer"
	"axj/Kt/KtUnsafe"
	"axjGW/pkg/gateway"
	"encoding/json"
	"golang.org/x/net/websocket"
)

type ProcessorExt struct {
	ANet.ProcessorV
}

func (that *ProcessorExt) ReqOpen(i int, pBuffer **KtBuffer.Buffer, conn ANet.Conn, decryKey []byte) (error, int32, string, int32, []byte) {
	if i == 0 {
		return nil, 0, "", 0, nil

	} else if i == 1 {
		if websocket, ok := conn.(*ANet.ConnWebsocket); ok {
			req := websocket.Conn().Request()
			data, err := json.Marshal([]interface{}{req.RequestURI, req.Header})
			//Kt.Log(KtUnsafe.BytesToString(data))
			return err, 0, "", 0, data
		}
	}

	return that.Req(pBuffer, conn, decryKey)
}

func (that *ProcessorExt) Req(pBuffer **KtBuffer.Buffer, conn ANet.Conn, decryKey []byte) (error, int32, string, int32, []byte) {
	err, bs, _ := conn.ReadA()
	return err, ANet.REQ_ONEWAY, "S", 0, bs
}

func (that *ProcessorExt) Rep(bufferP bool, conn ANet.Conn, encryKey []byte, compress bool, req int32, uri string, uriI int32, data []byte, isolate bool, id int64) error {
	if req < 0 {
		// websocket心跳
		if web, ok := conn.(*ANet.ConnWebsocket); ok {
			return websocket.Message.Send(web.Conn(), "")
		}

	} else if req > 0 {
		return nil
	}

	if uri != "" {
		if web, ok := conn.(*ANet.ConnWebsocket); ok {
			return websocket.Message.Send(web.Conn(), uri)
		}

		return conn.Write(KtUnsafe.StringToBytes(uri))
	}

	if data != nil {
		return conn.Write(data)
	}

	return nil
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
