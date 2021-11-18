package main

import (
	"axj/APro"
	"axj/Kt/KtCvt"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/Util"
	"axjGW/pkg/asdk"
	"bufio"
	"container/list"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Opt struct {
}

func (o Opt) LoginData(adapter *asdk.Adapter) []byte {
	return KtUnsafe.StringToBytes(strconv.FormatInt(time.Now().UnixNano(), 10))
}

func (o Opt) OnPush(uri string, data []byte, tid int64) {
	if State.log {
		fmt.Println("OnPush " + uri + ", " + strconv.FormatInt(tid, 10))
	}

	State.onRec(uri, data)
}

func (o Opt) OnLast(gid string, connVer int32, continues bool) {
	if State.log {
		fmt.Println("OnLast " + gid + ", " + strconv.Itoa(int(connVer)) + ", " + strconv.FormatBool(continues))
	}

	State.onRec(gid, nil)
}

func (o Opt) OnState(adapter *asdk.Adapter, state int, err string, data []byte) {
	if State.log || state == asdk.ERROR {
		fmt.Println("OnState , " + strconv.Itoa(state) + ", " + err)
	}

	State.onState(state, data)
}

func (o Opt) LoadStorage(name string) string {
	return ""
}

func (o Opt) SaveStorage(name string, value string) {
}

type clientsState struct {
	addr       string
	out        bool
	encry      bool
	log        bool
	locker     sync.Locker
	checkAsync *Util.NotifierAsync
	list       *list.List
	num        int   // 客户端总数
	loopedNum  int   // 已连接数
	errorNum   int   // 错误累计数
	closeNum   int   // 断开数
	recNum     int   // 收到消息数
	recDlyMin  int64 // 收到消息最小延迟
	recDlyMax  int64 // 收到消息最大延迟
	recDlyNum  int64 // 收到消息延迟数
	recDlyTime int64 // 收到消息总延迟
}

var State *clientsState

func init() {
	that := new(clientsState)
	// 127.0.0.1:8683
	// ws://127.0.0.1:8682/gw
	that.addr = "127.0.0.1:8683"
	that.log = false
	that.locker = new(sync.Mutex)
	that.checkAsync = Util.NewNotifierAsync(that.check, that.locker)
	that.list = new(list.List)
	State = that
	that.reset()
	go func() {
		for {
			// 定时检查
			time.Sleep(60 * time.Second)
			that.checkAsync.Start(nil)
		}
	}()
}

func (that *clientsState) reset() {
	that.locker.Lock()
	defer that.locker.Unlock()
	that.log = false
	that.out = true
	that.encry = true
	that.num = 0
	that.loopedNum = 0
	that.errorNum = 0
	that.closeNum = 0
	that.recNum = 0
	that.recDlyMin = 0
	that.recDlyMax = 0
	that.recNum = 0
	that.recDlyTime = 0
}

func (that *clientsState) clear() {
	that.locker.Lock()
	defer that.locker.Unlock()
	that.recNum = 0
	that.recDlyMin = 0
	that.recDlyMax = 0
	that.recNum = 0
	that.recDlyTime = 0
}

func (that *clientsState) next(el *list.Element) *list.Element {
	that.locker.Lock()
	defer that.locker.Unlock()
	if el == nil {
		return that.list.Front()
	}

	return el.Next()
}

func (that *clientsState) remove(el *list.Element) {
	that.locker.Lock()
	defer that.locker.Unlock()
	that.list.Remove(el)
}

func (that *clientsState) push(client *asdk.Client) {
	that.locker.Lock()
	defer that.locker.Unlock()
	that.list.PushBack(client)
}

func (that *clientsState) check() {
	tNum := that.num
	tAdapters := make([]*asdk.Adapter, tNum+1)
	el := that.next(nil)
	num := 0
	nxt := el
	for {
		el = nxt
		if el == nil {
			break
		}

		nxt = that.next(el)
		client := el.Value.(*asdk.Client)
		num++
		if num <= tNum {
			tAdapters[num] = client.Conn()

		} else {
			client.Close()
			that.remove(el)
		}
	}

	for num < tNum {
		client := asdk.NewClient(that.addr, that.out, that.encry, 256, 256<<10, 3, 8320, &Opt{})
		that.push(client)
		num++
		tAdapters[num] = client.Conn()
	}

	// loopedNum计算锁
	that.locker.Lock()
	defer that.locker.Unlock()
	loopedNum := 0
	for _, adapter := range tAdapters {
		if adapter != nil {
			if adapter.IsLooped() {
				loopedNum++
			}
		}
	}

	that.loopedNum = loopedNum
}

func (that *clientsState) setNum(num int) {
	that.num = num
	that.checkAsync.Start(nil)
}

func (that *clientsState) clientAny(looped bool) *asdk.Client {
	el := that.next(nil)
	nxt := el
	for {
		el = nxt
		if el == nil {
			break
		}

		nxt = that.next(el)
		client := el.Value.(*asdk.Client)
		adapter := client.Conn()
		if adapter != nil {
			if !looped || adapter.IsLooped() {
				return client
			}
		}
	}

	return nil
}

func (that *clientsState) send(uri string) {
	client := that.clientAny(false)
	if client == nil {
		fmt.Println("send clientAny not found")
		return
	}

	sTime := time.Now().UnixNano() / 1000000
	sTimeS := strconv.FormatInt(sTime, 10)
	jsonData, _ := json.Marshal([]string{uri, sTimeS})
	client.Req("test/sendU", jsonData, true, 30, func(s string, bytes []byte) {
		fmt.Printf("send rep %s span %dms, err %s\n", sTimeS, (time.Now().UnixNano()/1000000 - sTime), s)
	})
}

func (that *clientsState) onRec(uri string, data []byte) {
	var sTime int64 = 0
	if data != nil {
		str := KtUnsafe.BytesToString(data)
		sTime = KtCvt.ToInt64(str)
	}

	that.locker.Lock()
	defer that.locker.Unlock()
	that.recNum++
	if sTime > 0 {
		sTime = time.Now().UnixNano()/1000000 - sTime
		if that.recDlyMin == 0 || that.recDlyMin > sTime {
			that.recDlyMin = sTime
		}

		if that.recDlyMax < sTime {
			that.recDlyMax = sTime
		}

		that.recDlyNum++
		that.recDlyTime += sTime
	}
}

func (that *clientsState) onState(state int, data []byte) {
	that.locker.Lock()
	defer that.locker.Unlock()
	switch state {
	case asdk.LOOP:
		that.loopedNum++
		break
	case asdk.ERROR:
		that.errorNum++
		break
	case asdk.CLOSE:
		that.closeNum++
		break
	}
}

func (that *clientsState) printStatus() {
	fmt.Printf("addr: %s\n", that.addr)
	fmt.Printf("log: %v\n", that.log)
	fmt.Printf("num: %d\n", that.num)
	fmt.Printf("loopedNum: %d\n", that.loopedNum)
	fmt.Printf("errorNum: %d\n", that.errorNum)
	fmt.Printf("closeNum: %d\n", that.closeNum)
	fmt.Printf("recNum: %d\n", that.recNum)
	fmt.Printf("recDlyMin: %dms\n", that.recDlyMin)
	fmt.Printf("recDlyMax: %dms\n", that.recDlyMax)
	fmt.Printf("recDlyNum: %d\n", that.recDlyNum)
	var recDlyAvg int64 = 0
	recDlyNum := that.recDlyNum
	if recDlyNum > 0 {
		recDlyAvg = that.recDlyTime / recDlyNum
	}

	fmt.Printf("recDlyAvg: %dms\n", recDlyAvg)
}

type cmder struct {
	cmd  string
	help string
	fun  func(str string)
}

func main() {
	cmders := []cmder{
		{
			cmd:  "addr",
			help: "addr $addr//设置客户端连接地址",
			fun: func(str string) {
				State.addr = str
			},
		},
		{
			cmd:  "log",
			help: "log $log//设置客户端打印日志",
			fun: func(str string) {
				State.log = KtCvt.ToBool(str)
			},
		},
		{
			cmd:  "out",
			help: "out $out//设置客户端协议Out流写入",
			fun: func(str string) {
				State.out = KtCvt.ToBool(str)
			},
		},
		{
			cmd:  "encry",
			help: "encry $encry//设置客户端通讯加密",
			fun: func(str string) {
				State.encry = KtCvt.ToBool(str)
			},
		},
		{
			cmd:  "reset",
			help: "reset//重置客户端状态",
			fun: func(str string) {
				State.reset()
			},
		},
		{
			cmd:  "clear",
			help: "clear//清除客户端延迟状态",
			fun: func(str string) {
				State.clear()
			},
		},
		{
			cmd:  "status",
			help: "status//打印客户端状态",
			fun: func(str string) {
				State.printStatus()
			},
		},
		{
			cmd:  "client",
			help: "client $num//设置客户端数量",
			fun: func(str string) {
				State.setNum(int(KtCvt.ToInt32(str)))
			},
		},
		{
			cmd:  "send",
			help: "send $uri//发送测试消息",
			fun: func(str string) {
				State.send(str)
			},
		},
		{
			cmd:  "check",
			help: "check //客户端列表检查",
			fun: func(str string) {
				State.checkAsync.Start(nil)
			},
		},
	}

	go func() {
		input := bufio.NewScanner(os.Stdin)
		// 逐行扫描
		for input.Scan() && !APro.Stopped {
			line := strings.TrimSpace(input.Text())
			if line == "" {
				break
			}

			cmd := line
			data := ""
			idx := strings.IndexByte(line, ' ')
			if idx > 0 {
				cmd = strings.ToLower(line[0:idx])
				data = strings.TrimSpace(line[idx+1:])

			} else {
				cmd = strings.ToLower(line)
			}

			did := false
			for _, cer := range cmders {
				if cer.cmd == cmd {
					did = true
					cer.fun(data)
					break
				}
			}

			if did {
				fmt.Println(cmd + "[" + data + "] done")

			} else {
				if cmd == "help" || cmd == "-help" || cmd == "--help" {
					for _, cer := range cmders {
						fmt.Println(cer.help)
					}

				} else {
					fmt.Println("cmder not found for " + cmd)
				}
			}
		}
	}()

	APro.Signal()
}
