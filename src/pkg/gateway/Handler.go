package gateway

import (
	"axj/ANet"
	"axj/APro"
)

var Processor = ANet.Processor{
	Protocol: ANet.ProtocolV{},
	Compress: ANet.CompressZip{},
	CompressMin: APro.Cfg,
	Encrypt:  ANet.EncryptSr{},
}

func init() {

}

type Handler struct {
}
