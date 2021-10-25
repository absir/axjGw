package ANet

import (
	"axj/Kt/Kt"
	"axj/Thrd/Util"
	"sync"
	"time"
)

type ClientM interface {
	Client
	GetM() *ClientMng
}

type ClientMng struct {
	ClientCnn
	id       int64
	initTime int64
	idleTime int64
}

func (that *ClientMng) GetM() *ClientMng {
	return that
}

func (that ClientMng) Id() int64 {
	return that.id
}

func (that ClientMng) InitTime() int64 {
	return that.initTime
}

type HandlerM interface {
	Handler
	New(conn Conn) ClientM
	Check(time int64, client Client) // nio
}

type Manager struct {
	handlerM  HandlerM
	clientMap *sync.Map
	idWorker  *Util.IdWorker
	idleDrt   int64
	checkDrt  time.Duration
	checkLoop int64
	beatBytes []byte
}

func (that Manager) HandlerM() HandlerM {
	return that.handlerM
}

func (that Manager) ClientMap() *sync.Map {
	return that.clientMap
}

func (that Manager) IdWorker() *Util.IdWorker {
	return that.idWorker
}

func (that Manager) Client(cid int64) Client {
	client, ok := that.clientMap.Load(cid)
	if ok {
		return client.(Client)
	}

	return nil
}

func NewManager(handlerM HandlerM, workerId int32, idleDrt time.Duration, checkDrt time.Duration) *Manager {
	idWorker, err := Util.NewIdWorker(workerId)
	Kt.Panic(err)
	that := new(Manager)
	that.handlerM = handlerM
	that.clientMap = new(sync.Map)
	that.checkDrt = checkDrt
	that.idleDrt = int64(idleDrt)
	that.idWorker = idWorker
	that.beatBytes = handlerM.Processor().Protocol.Rep(REQ_BEAT, "", 0, nil, false, 0)
	return that
}

func (that Manager) ClientM(client Client) *ClientMng {
	return client.(ClientM).GetM()
}

func (that Manager) OnOpen(client Client) {
	clientM := that.ClientM(client)
	if clientM.id <= 0 {
		clientM.id = that.idWorker.Generate()
	}

	that.clientMap.Store(clientM.id, client)
	that.handlerM.OnOpen(client)
}

func (that Manager) OnClose(client Client, err error, reason interface{}) {
	// Map删除
	that.clientMap.Delete(that.ClientM(client).id)
	that.handlerM.OnClose(client, err, reason)
}

func (that Manager) OnKeep(client Client, req bool) {
	that.ClientM(client).idleTime = time.Now().UnixNano() + that.idleDrt
	that.handlerM.OnKeep(client, req)
}

func (that Manager) OnReq(client Client, req int32, uri string, uriI int32, data []byte) bool {
	if req == REQ_BEAT {
		return true
	}

	return that.handlerM.OnReq(client, req, uri, uriI, data)
}

func (that Manager) OnReqIO(client Client, req int32, uri string, uriI int32, data []byte) {
	that.handlerM.OnReqIO(client, req, uri, uriI, data)
}

func (that Manager) Processor() Processor {
	return that.handlerM.Processor()
}

func (that Manager) UriDict() UriDict {
	return that.handlerM.UriDict()
}

// 空闲检测
func (that Manager) CheckStop() {
	that.checkLoop = -1
}

func (that Manager) CheckLoop() {
	loopTime := time.Now().UnixNano()
	that.checkLoop = loopTime
	for loopTime == that.checkLoop {
		time.Sleep(that.checkDrt)
		time := time.Now().UnixNano()
		that.clientMap.Range(func(key, value interface{}) bool {
			that.checkClient(time, key, value.(Client))
			return true
		})
	}
}

func (that Manager) checkClient(time int64, key interface{}, client Client) {
	clientM, _ := client.(ClientM)
	if clientM == nil {
		that.clientMap.Delete(key)
		return
	}

	clientC := client.Get()
	// 已关闭链接
	if client.Get().IsClosed() {
		that.clientMap.Delete(key)
		return
	}

	if clientM.GetM().idleTime <= time {
		// 直接心跳
		that.OnKeep(client, false)
		go clientC.Rep(true, -1, "", 0, that.beatBytes, false, false)
	}

	that.handlerM.Check(time, client)
}

func (that Manager) Open(conn Conn, encryKey []byte, id int64) Client {
	handlerM := that.handlerM
	client := handlerM.New(conn)
	clientM := that.ClientM(client)
	clientM.id = id
	clientM.initTime = time.Now().UnixNano()
	clientM.idleTime = clientM.initTime + that.idleDrt
	client.Get().Open(conn, that, encryKey)
	handlerM.OnOpen(client)
	return client
}
