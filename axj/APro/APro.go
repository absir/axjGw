package APro

import (
	Kt2 "axj/Kt/Kt"
	KtCfg2 "axj/Kt/KtCfg"
	KtCvt2 "axj/Kt/KtCvt"
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
		Kt2.Err(errors.New("path has got"), false)
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
		Kt2.Err(err, true)
		tmp = dir
	}

	return tmp
}

func Path() string {
	if path == "" {
		file, err := os.Executable()
		Kt2.Err(err, true)
		if file == "" || (strings.HasPrefix(file, Tmp()) && callerP != nil) {
			if file != "" {
				Kt2.Log("exe : " + file)
				Kt2.Log("tmp : " + Tmp())
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
			Kt2.Err(err, true)
			if file != "" {
				path = file
				callerP = nil
			}
		}
	}

	return path
}

var Cfg KtCfg2.Cfg = nil

func Load(reader *bufio.Reader, entry string) KtCfg2.Cfg {
	if Cfg == nil {
		readMap := map[string]KtCfg2.Read{}
		readMap["@env"] = func(str string) {
			str = strings.ToLower(str)
			switch str {
			case "dev":
			case "develop":
				Kt2.Env = Kt2.Develop
				break
			case "test":
				Kt2.Env = Kt2.Test
				break
			case "debug":
				Kt2.Env = Kt2.Debug
				break
			case "pro":
			case "prod":
			case "product":
				Kt2.Env = Kt2.Product
				break
			default:
				Kt2.Env = KtCvt2.ToType(str, KtCvt2.Int8).(int8)
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
			Kt2.Err(err, true)
			if str != "" {
				if !loads[str] {
					loads[str] = true
					cfgs.PushBack(str)
				}
			}
		}

		_cfg := KtCfg2.Cfg{}
		if reader != nil {
			_cfg = KtCfg2.ReadIn(reader, _cfg, &readMap).(KtCfg2.Cfg)
		}

		readMap["@cfg"](entry)
		for {
			el := cfgs.Front()
			if el == nil {
				break
			}

			f, err := os.Open(cfgs.Remove(el).(string))
			Kt2.Err(err, true)
			if f != nil {
				_cfg = KtCfg2.ReadIn(bufio.NewReader(f), _cfg, &readMap).(KtCfg2.Cfg)

			} else {
				break
			}
		}

		var fun *KtCfg2.Read = nil
		name := ""
		var lst *list.List = nil
		args := os.Args
		lenA := len(args)
		for i := 1; i < lenA; i++ {
			arg := args[i]
			if arg[0] == '-' {
				if strings.IndexByte(arg, '=') > 0 {
					if fun == nil {
						f := KtCfg2.ReadFunc(_cfg, &readMap)
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
						f := KtCfg2.ReadFunc(_cfg, &readMap)
						fun = &f
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
func SubCfg(sub string) Kt2.Map {
	var cfg Kt2.Map = nil
	m := Kt2.If(Cfg == nil, nil, KtCfg2.Get(cfg, sub))
	if m != nil {
		mp, ok := m.(*Kt2.Map)
		if ok && mp != nil {
			cfg = *mp
		}
	}

	env := os.Getenv(sub + "_CFG")
	if env != "" {
		cfg = KtCfg2.ReadIn(bufio.NewReader(strings.NewReader(env)), cfg, nil)
	}

	return cfg
}

func SubCfgBind(sub string, bind interface{}) interface{} {
	mp := SubCfg(sub)
	KtCvt2.BindInterface(bind, mp)
	return bind
}

// 关闭信号
func Signal() os.Signal {
	c := make(chan os.Signal, 0)
	signal.Notify(c, syscall.SIGTERM)
	return <-c
}
