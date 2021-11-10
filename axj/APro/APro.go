package APro

import (
	"axj/Kt/Kt"
	"axj/Kt/KtCfg"
	"axj/Kt/KtCvt"
	"axj/Kt/KtStr"
	"bufio"
	"container/list"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
)

type CallerP struct {
	fun func(skip int) (pc uintptr, file string, line int, ok bool)
	dir string
}

var callerP *CallerP

var tmp = ""
var path = ""
var Locker = new(sync.Mutex)

/*
APro.Caller(func(skip int) (pc uintptr, file string, line int, ok bool) {
		return runtime.Caller(0)
	}, "../public")
*/
func Caller(fun func(skip int) (pc uintptr, file string, line int, ok bool), dir string) {
	if path != "" {
		Kt.Err(errors.New("path has got"), false)
		return
	}

	callerP = &CallerP{fun, dir}
}

// 获取系统临时目录，兼容go run
func Tmp() string {
	if tmp == "" {
		dir, err := ioutil.TempDir("", "")
		if dir == "" || err != nil {
			dir = os.Getenv("TEMP")
			if dir == "" {
				dir = os.Getenv("TMP")
				if dir == "" {
				}
			}

		} else {
			dir = filepath.Dir(dir)
		}

		dir, err = filepath.EvalSymlinks(dir)
		Kt.Err(err, true)
		tmp = dir
	}

	return tmp
}

func atTmpFile(file string) bool {
	idx := strings.Index(Tmp(), file)
	if idx < 0 {
		return false
	}

	if idx == 0 {
		return true
	}

	if len(file) > 0 && file[0] == '/' {
		prv := file[1:idx]
		switch prv {
		case "private":
			return true
		}
	}

	return false
}

func Path() string {
	if path == "" {
		file, err := os.Executable()
		Kt.Err(err, true)
		Kt.Log("exe : " + file)
		Kt.Log("tmp : " + Tmp())
		if file == "" && callerP != nil || atTmpFile(file) {
			file = ""
			ok := true
			if callerP == nil {
				_, file, _, ok = runtime.Caller(0)
				if ok {
					file = filepath.Dir(file)
				}

			} else {
				_, file, _, ok = callerP.fun(0)
				if ok {
					file = filepath.Dir(file)
					if callerP.dir != "" {
						file = filepath.Join(file, callerP.dir)
					}
				}
			}

		} else {
			file = filepath.Dir(file)
		}

		if file != "" {
			file, err = filepath.EvalSymlinks(file)
			Kt.Err(err, true)
			if file != "" {
				path = file
				callerP = nil
			}
		}

		Kt.Log("path : " + path)
	}

	return path
}

var Cfg KtCfg.Cfg = nil

func Load(reader *bufio.Reader, entry string) KtCfg.Cfg {
	if Cfg == nil {
		readMap := map[string]KtCfg.Read{}
		readMap["@env"] = func(str string) {
			str = strings.ToLower(str)
			switch str {
			case "dev":
			case "develop":
				Kt.Env = Kt.Develop
				break
			case "test":
				Kt.Env = Kt.Test
				break
			case "debug":
				Kt.Env = Kt.Debug
				break
			case "pro":
			case "prod":
			case "product":
				Kt.Env = Kt.Product
				break
			default:
				Kt.Env = KtCvt.ToType(str, KtCvt.Int8).(int8)
				break
			}
		}

		loads := map[string]bool{}
		cfgs := list.New()
		readMap["@cfg"] = func(str string) {
			if str == "" {
				return
			}

			str, err := filepath.EvalSymlinks(filepath.Join(Path(), str))
			Kt.Err(err, true)
			if str != "" {
				if !loads[str] {
					loads[str] = true
					cfgs.PushBack(str)
				}
			}
		}

		_cfg := KtCfg.Cfg{}
		if reader != nil {
			_cfg = KtCfg.ReadIn(reader, _cfg, &readMap).(KtCfg.Cfg)
		}

		readMap["@cfg"](entry)
		for {
			el := cfgs.Front()
			if el == nil {
				break
			}

			f, err := os.Open(cfgs.Remove(el).(string))
			Kt.Err(err, true)
			if f != nil {
				_cfg = KtCfg.ReadIn(bufio.NewReader(f), _cfg, &readMap).(KtCfg.Cfg)

			} else {
				break
			}
		}

		var fun *KtCfg.Read = nil
		name := ""
		var lst *list.List = nil
		args := os.Args
		lenA := len(args)
		for i := 1; i < lenA; i++ {
			arg := args[i]
			if arg[0] == '-' {
				if strings.IndexByte(arg, '=') > 0 {
					if fun == nil {
						f := KtCfg.ReadFunc(_cfg, &readMap)
						*fun = f
					}

					(*fun)(arg[1:])

				} else {
					name = arg[1:]
				}

			} else {
				if name == "" {
					// 顺序参数
					if lst == nil {
						lst = list.New()
					}

					lst.PushBack(arg)

				} else {
					// 分离参数
					if fun == nil {
						f := KtCfg.ReadFunc(_cfg, &readMap)
						*fun = f
					}

					(*fun)(name + "=" + arg)
					name = ""
				}
			}
		}

		if lst != nil {
			_cfg[":args"] = lst
		}

		Cfg = _cfg
	}

	return Cfg
}

// 子配置
func SubCfg(sub string) Kt.Map {
	m := Kt.If(Cfg == nil, nil, KtCfg.Get(Cfg, sub))
	var cfg Kt.Map = nil
	if m != nil {
		mp, ok := m.(Kt.Map)
		if ok && mp != nil {
			cfg = mp
		}
	}

	env := os.Getenv(sub + "_CFG")
	if env != "" {
		cfg = KtCfg.ReadIn(bufio.NewReader(strings.NewReader(env)), cfg, nil)
	}

	return cfg
}

func SubCfgBind(sub string, bind interface{}) interface{} {
	mp := SubCfg(sub)
	KtCvt.BindInterface(bind, mp)
	return bind
}

func FileCfg(file string) KtCfg.Cfg {
	file = filepath.Join(Path(), file)
	f, err := os.Open(file)
	if os.IsNotExist(err) {
		return nil
	}

	Kt.Err(err, true)
	if f != nil {
		return KtCfg.ReadIn(bufio.NewReader(f), nil, nil).(KtCfg.Cfg)

	} else {
		return nil
	}
}

var workId int32 = -1

func WorkId() int32 {
	if workId < 0 {
		Locker.Lock()
		defer Locker.Unlock()
		if workId < 0 {
			var id int32 = 0
			if Cfg != nil {
				id = KtCfg.GetType(Cfg, "workId", KtCvt.Int32, 0, "").(int32)
			}

			if id <= 0 {
				str := os.Getenv("HOSTNAME")
				i := strings.LastIndexByte(str, '-')
				if i > 0 {
					str = str[i+1:]
				}

				if KtStr.DigitStr(str) {
					id = KtCvt.ToType(str, KtCvt.Int32).(int32)
				}

				if id < 0 {
					id = 0
				}
			}

			workId = id
		}
	}

	return workId
}

// 关闭信号
func Signal() os.Signal {
	c := make(chan os.Signal, 0)
	signal.Notify(c, syscall.SIGTERM)
	s := <-c
	fmt.Printf("exit pro ------- signal:[%v]", s)
	return s
}
