package nps

import (
	"axj/APro"
	"axj/Kt/KtEncry"
	"axj/Kt/KtRand"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/cmap"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var RandChars = KtUnsafe.StringToBytes("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

var loginToken = ""

func NpsApiInit() {
	// 登录
	if len(NpsConfig.AdminAddr) <= 1 || NpsConfig.AdminName == "" || NpsConfig.AdminPassword == "" {
		return
	}

	passwordHash := KtEncry.EnMd5(KtUnsafe.StringToBytes(NpsConfig.AdminPassword))
	http.HandleFunc("/login", func(writer http.ResponseWriter, request *http.Request) {
		username := request.FormValue("username")
		password := request.FormValue("password")
		if username == NpsConfig.AdminName && password == passwordHash {
			loginToken = KtRand.RandString(16, RandChars)
			cookie := http.Cookie{
				Name:    "atoken",
				Value:   loginToken,
				Expires: time.Now().AddDate(0, 0, 1), // 过期时间为当前时间加一天
			}
			http.SetCookie(writer, &cookie)
			fmt.Fprintf(writer, "ok") // 返回"ok"消息

		} else {
			fmt.Fprintf(writer, "fail") // 返回"fail"消息
		}
	})

	mapForName := func(name string) *cmap.CMap {
		if "client" == name {
			return ClientMap
		} else if "host" == name {
			return HostMap
		} else if "tcp" == name {
			return TcpMap
		}

		return nil
	}

	regLoginedFunc("/list/", func(writer http.ResponseWriter, request *http.Request) {
		path := request.URL.Path
		name := path[len("/list/"):]
		bs, _ := json.Marshal(ReadList(mapForName(name), true))
		writer.Write(bs)
	})

	editLocker := new(sync.Mutex)
	regLoginedFunc("/del/", func(writer http.ResponseWriter, request *http.Request) {
		path := request.URL.Path
		parts := strings.Split(path[len("/del/"):], "/")
		if len(parts) < 2 {
			http.Error(writer, "Invalid request", http.StatusBadRequest)
			return
		}

		name := parts[0]
		id, _ := strconv.Atoi(parts[1])

		// 编辑单进程
		editLocker.Lock()
		defer editLocker.Unlock()
		MapDel(mapForName(name), id)
		fmt.Fprintf(writer, "ok") // 返回"ok"消息
	})

	regLoginedFunc("/edit/", func(writer http.ResponseWriter, request *http.Request) {
		path := request.URL.Path
		name := path[len("/list/"):]
		cmap := mapForName(name)
		if cmap != nil {
			var npsId NpsId = nil
			if "client" == name {
				npsId = &NpsClient{}
			} else if "host" == name {
				npsId = &NpsHost{}
			} else if "tcp" == name {
				npsId = &NpsTcp{}
			}

			bs, _ := io.ReadAll(request.Body)
			json.Unmarshal(bs, npsId)

			// 编辑单进程
			editLocker.Lock()
			defer editLocker.Unlock()
			if npsId.GetId() <= 0 {
				id := int(cmap.Count()) + 1
				for {
					if _, ok := cmap.Load(id); !ok {
						break
					}

					id++
				}

				npsId.SetId(id)
			}

			old, _ := cmap.Load(npsId.GetId())
			cmap.Store(npsId.GetId(), npsId)
			MapDirty(cmap, npsId, old, false)
			fmt.Fprintf(writer, "ok") // 返回"ok"消息
			return
		}

		fmt.Fprintf(writer, "fail") // 返回"fail"消息
	})

	// web文件
	web := http.FileServer(http.Dir(filepath.Join(APro.Path(), "web")))
	http.Handle("/web/", http.StripPrefix("/web/", web))

	// 面板http服务
	http.ListenAndServe(NpsConfig.AdminAddr, nil)
}

func regLoginedFunc(pattern string, handler func(writer http.ResponseWriter, request *http.Request)) {
	http.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		token, _ := request.Cookie("atoken")
		if token == nil || token.Value != loginToken {
			fmt.Fprintf(writer, "noLogin") // 返回"ok"消息
			return
		}

		if handler != nil {
			writer.Header().Set("Content-Type", "application/json; charset=utf-8") // 设置响应头，指定Content-Type和字符集为UTF-8
			handler(writer, request)
		}
	})
}
