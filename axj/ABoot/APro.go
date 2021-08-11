package ABoot

import (
	"axj/Kt"
	"axj/KtCfg"
	"axj/KtCvt"
	"bufio"
	"container/list"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

var path = ""

func Path() string {
	if path == "" {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			Kt.Err(err)
		}

		path = strings.Replace(dir, "\\", "/", -1)
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
		cfgs := new(list.List)
		readMap["@cfg"] = func(str string) {
			if !loads[str] {
				loads[str] = true
				cfgs.PushBack(str)
			}
		}

		_cfg := new(KtCfg.Cfg)
		if reader != nil {
			_cfg = KtCfg.ReadIn(reader, _cfg, &readMap)
		}

		readMap["@cfg"](entry)
		for {
			el := cfgs.Front()
			if el == nil {
				break
			}

			f, err := os.Open(el.Value.(string))
			Kt.Err(err)
			if f != nil {
				_cfg = KtCfg.ReadIn(bufio.NewReader(f), _cfg, &readMap)
			}
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
