package APro

import (
	"axj/Kt/Kt"
	"axj/Kt/KtCfg"
	"axj/Kt/KtCvt"
	"axj/Kt/KtFile"
	"axj/Kt/KtStr"
	"bufio"
	"container/list"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
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
	tmp := Tmp()
	if strings.HasPrefix(file, tmp) {
		return true
	}

	if len(tmp) > 0 && tmp[0] == '/' {
		idx := KtStr.IndexByte(tmp, '/', 1)
		if idx > 0 {
			prv := tmp[1:idx]
			switch prv {
			case "private":
				break
			default:
				return false
			}

			return strings.HasPrefix(file, tmp[idx:])
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
		// 配置读取处理
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

		// 入口文件参数配置
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

		// 环境变量参数配置
		acfg := os.Getenv("_acfg")
		if acfg != "" {
			_cfg = KtCfg.ReadIn(bufio.NewReader(strings.NewReader(acfg)), _cfg, &readMap).(KtCfg.Cfg)
		}

		// 命令行参数配置
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

func Open(file string) *os.File {
	return KtFile.Open(filepath.Join(Path(), file))
}

func Create(file string, append bool) *os.File {
	return KtFile.Create(filepath.Join(Path(), file), append)
}

func FileCfg(file string) KtCfg.Cfg {
	f := Open(file)
	if f == nil {
		return nil
	}

	return KtCfg.ReadIn(bufio.NewReader(f), nil, nil).(KtCfg.Cfg)
}

func GetCfg(name string, typ reflect.Type, dVal interface{}) interface{} {
	return KtCfg.GetType(Cfg, name, typ, dVal)
}

var workId int32 = -1

func WorkId() int32 {
	if workId < 0 {
		Locker.Lock()
		defer Locker.Unlock()
		if workId < 0 {
			var id int32 = 0
			if Cfg != nil {
				id = KtCfg.GetType(Cfg, "workId", KtCvt.Int32, 0).(int32)
			}

			if id <= 0 {
				str := os.Getenv("HOSTNAME")
				i := strings.LastIndexByte(str, '-')
				if i > 0 {
					str = str[i+1:]
				}

				if KtStr.DigitStr(str) {
					id = KtCvt.ToInt32(str)
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

var startRuns []func()

func StartAdd(run func()) {
	if Kt.Started {
		panic("APro Is Started")
	}

	if run == nil {
		return
	}

	if startRuns == nil {
		startRuns = []func(){run}

	} else {
		startRuns = append(startRuns, run)
	}
}

func Start() {
	if Kt.Started {
		return
	}

	if stopRuns != nil {
		for _, run := range startRuns {
			run()
		}

		startRuns = nil
	}

	Kt.Started = true
	Kt.Log("APro Started")
}

var Stopped bool

var stopRuns []func()

func StopAdd(run func()) {
	if Stopped {
		panic("APro Is Stopped")
	}

	if run == nil {
		return
	}

	if stopRuns == nil {
		stopRuns = []func(){run}

	} else {
		stopRuns = append(stopRuns, run)
	}
}

func Stop() {
	if Stopped {
		return
	}

	Kt.Active = false
	if stopRuns != nil {
		for _, run := range stopRuns {
			run()
		}

		stopRuns = nil
	}

	Stopped = true
	Kt.Log("APro Stopped")
}

// 开启关闭信号
func Signal() {
	Start()
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGILL)
	s := <-c
	Kt.Log(fmt.Sprintf("os signal %v", s))
	defer Exit(0)
	Stop()
}

func Exit(code int) {
	Kt.Log(fmt.Sprintf("os exit %d", code))
	os.Exit(code)
}
