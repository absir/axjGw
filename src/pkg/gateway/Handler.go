package gateway

import "axj/ANet"

var Processor = ANet.Processor{
	Protocol: ANet.ProtocolV{},
	Compress: ANet.CompressZip{},

	Encrypt:  ANet.EncryptSr{},
}

func init() {

}

type Handler struct {
}
