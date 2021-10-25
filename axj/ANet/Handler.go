package ANet

type Handler interface {
	OnOpen(client Client)
	OnClose(client Client, err error, reason interface{})
	OnKeep(client Client, req bool) // 保持连接
	OnReq(client Client, req int32, uri string, uriI int32, data []byte) bool
	OnReqIO(client Client, req int32, uri string, uriI int32, data []byte)
	Processor() Processor
	UriDict() UriDict
}
