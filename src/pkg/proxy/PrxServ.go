package proxy

import "axjGW/gen/gw"

type prxServ struct {
}

var PrxServ = new(prxServ)

func (that *prxServ) Init(workId int32, cfg map[interface{}]interface{}, aclClient gw.AclClient) {

}

func (that *prxServ) Start() {

}
