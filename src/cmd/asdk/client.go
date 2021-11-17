package main

import (
	"axj/APro"
	"axj/Kt/KtUnsafe"
	"axjGW/pkg/asdk"
	"fmt"
	"strconv"
	"time"
)

type Opt struct {
}

func (o Opt) LoginData(adapter *asdk.Adapter) []byte {
	return KtUnsafe.StringToBytes(strconv.FormatInt(time.Now().UnixNano(), 10))
}

func (o Opt) OnPush(uri string, data []byte, tid int64) {
	fmt.Println("OnPush " + uri + ", " + strconv.FormatInt(tid, 10))
}

func (o Opt) OnLast(gid string, connVer int32, continues bool) {
	fmt.Println("OnLast " + gid + ", " + strconv.Itoa(int(connVer)) + ", " + strconv.FormatBool(continues))
}

func (o Opt) OnState(adapter *asdk.Adapter, state int, err string, data []byte) {
	fmt.Println("OnState , " + strconv.Itoa(state) + ", " + err)
}

func (o Opt) LoadStorage(name string) string {
	return ""
}

func (o Opt) SaveStorage(name string, value string) {
}

func main() {
	// 127.0.0.1:8683
	// ws://127.0.0.1:8682/gw
	client := asdk.NewClient("127.0.0.1:8683", true, true, 256, 256<<10, 3, 8320, &Opt{})
	client.Conn()
	//for i := 0; i < 10; i++ {
	//	strs := [2]string{"uri" + strconv.Itoa(i), "data" + strconv.Itoa(i)}
	//	data, _ := json.Marshal(strs)
	//	client.Req("test/sendU", data, false, 30, func(s string, data []byte) {
	//		fmt.Println("send Back === " + s + KtUnsafe.BytesToString(data))
	//	})
	//}

	for i := 0; i < 2000; i++ {
		go func() {
			client := asdk.NewClient("127.0.0.1:8683", true, true, 256, 256<<10, 3, 8320, &Opt{})
			client.Conn()
		}()
	}

	APro.Signal()
}
