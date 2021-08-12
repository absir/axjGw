package APro

import (
	"axj/Kt"
	"axj/KtCfg"
	"axj/KtCvt"
	"bufio"
	"container/list"
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

type CallerP struct {
	fun func(skip int) (pc uintptr, file string, line int, ok bool)
	dir string
}

var callerP *CallerP

var tmp = ""
var path = ""

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
		dir := os.Getenv("TEMP")
		if dir == "" {
			dir = os.Getenv("TMP")
		}

		dir, err := filepath.EvalSymlinks(dir)
		Kt.Err(err, true)
		tmp = dir
	}

	return tmp
}

func Path() string {
	if path == "" {
		file, err := os.Executable()
		Kt.Err(err, true)
		if file == "" || (strings.HasPrefix(file, Tmp()) && callerP != nil) {
			if file != "" {
				Kt.Log("exe : " + file)
				Kt.Log("tmp : " + Tmp())
			}

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
	}

	return path
}

var cfg *KtCfg.Cfg = nil

func Cfg(reader *bufio.Reader, entry string) KtCfg.Cfg {
	if cfg == nil {
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
			str, err := filepath.EvalSymlinks(filepath.Join(Path(), str))
			Kt.Err(err, true)
			if str != "" {
				if !loads[str] {
					loads[str] = true
					cfgs.PushBack(str)
				}
			}
		}

		_cfg := &KtCfg.Cfg{}
		if reader != nil {
			_cfg = KtCfg.ReadIn(reader, _cfg, &readMap)
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
				_cfg = KtCfg.ReadIn(bufio.NewReader(f), _cfg, &readMap)

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
						f := KtCfg.ReadFunc(*_cfg, &readMap)
						fun = &f
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
						f := KtCfg.ReadFunc(*_cfg, &readMap)
						fun = &f
					}

					(*fun)(name + "=" + arg)
					name = ""
				}
			}
		}

		if lst != nil {
			(*_cfg)[":args"] = lst
		}

		cfg = _cfg
	}

	return *cfg
}

// 关闭信号
func Signal() os.Signal {
	c := make(chan os.Signal, 0)
	signal.Notify(c, syscall.SIGTERM)
	return <-c
}
