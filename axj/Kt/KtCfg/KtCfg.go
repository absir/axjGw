package KtCfg

import (
	"axj/Kt/Kt"
	"axj/Kt/KtCvt"
	"axj/Kt/KtStr"
	"bufio"
	"container/list"
	"io"
	"os"
	"reflect"
	"strings"
)

// 配置字典别名
type Cfg map[interface{}]interface{}

func (that Cfg) Map() map[interface{}]interface{} {
	return that
}

func (that Cfg) Range(fun func(key interface{}, val interface{}) bool) {
	if fun == nil {
		return
	}

	for key, val := range that {
		if !fun(key, val) {
			break
		}
	}
}

func (that Cfg) Get(key interface{}) interface{} {
	return that[key]
}

func (that Cfg) Put(key interface{}, val interface{}) interface{} {
	_val := that[key]
	that[key] = val
	return _val
}

func (that Cfg) Remove(key interface{}) interface{} {
	_val := that[key]
	delete(that, key)
	return _val
}

// 配置获取
func Get(cfg Kt.Map, name string) interface{} {
	val := Kt.If(cfg == nil, nil, cfg.Get(name))
	if val == nil {
		if len(name) > 1 {
			chr := name[0]
			switch chr {
			case '%':
				return os.Getenv(name[1:])
				break
			case '@':
			case '$':
				return Kt.If(cfg == nil, nil, cfg.Get(name[1:]))
				break
			}
		}

		if cfg != nil {
			ei := KtStr.IndexByte(name, '.', 0)
			if ei > 0 {
				mp := cfg
				si := 0
				for {
					if ei <= 0 {
						break
					}

					found := false
					if val = mp.Get(name[si:ei]); val != nil {
						if m, ok := val.(Kt.Map); ok {
							found = true
							mp = m
						}
					}

					if !found || mp == nil {
						return nil
					}

					si = ei + 1
					ei = KtStr.IndexByte(name, '.', si)
				}

				if mp == nil {
					return nil
				}

				return mp.Get(name[si:])
			}
		}
	}

	return val
}

func GetType(cfg Kt.Map, name string, typ reflect.Type, dVal interface{}) interface{} {
	val := Get(cfg, name)
	if val != nil {
		if typ == nil {
			if _, is := val.(*list.List); !is {
				if reflect.TypeOf(val).Kind() != reflect.Array {
					if _, is := val.(*list.List); !is {
						val = []interface{}{val}
					}
				}

				ary := KtCvt.ToType(Kt.If(val == nil, dVal, val), typ).([]interface{})
				val = Kt.ToList(ary)
				cfg.Put(name, ary)
			}

			return val

		} else {
			switch typ.Kind() {
			case reflect.Array:
				if reflect.TypeOf(val).Kind() != reflect.Array {
					if _, is := val.(*list.List); !is {
						val = []interface{}{val}
					}
				}
				break
			default:
				break
			}
		}
	}

	return KtCvt.ToType(Kt.If(val == nil, dVal, val), typ)
}

func GetExp(cfg Kt.Map, exp string, strict bool) string {
	return GetExpRemain(cfg, exp, strict, false)
}

func GetExpRemain(cfg Kt.Map, exp string, strict bool, remain bool) string {
	index := strings.Index(exp, "${")
	length := len(exp)
	if index >= 0 && index < length-2 {
		sb := &strings.Builder{}
		sIndex := 0
		for {
			if index > sIndex {
				sb.WriteString(exp[sIndex:index])

			} else if index < sIndex {
				if index < 0 {
					if length > sIndex {
						sb.WriteString(exp[sIndex:length])
					}
				}

				break
			}

			sIndex = KtStr.Index(exp, "${", index)
			if sIndex < 0 {
				sb.WriteString(exp[index:length])
				break
			}

			index += 2
			if index < sIndex {
				val := exp[index:sIndex]
				var valD interface{} = nil
				var value interface{} = nil
				if strings.IndexByte(val, '?') > 0 {
					// 支持二运运算|三元运算
					vals := KtStr.SplitStr(val, "?:", true, 0)
					val := vals[0]
					if len(vals) == 3 {
						strict = false
						valD = Kt.If(GetType(cfg, val.(string), KtCvt.Bool, false).(bool), vals[1], vals[2])

					} else {
						value = Get(cfg, val.(string))
						valD = vals[1]
					}

				} else {
					value = Get(cfg, val)
				}

				if value == nil {
					if valD != nil {
						sb.WriteString(valD.(string))

					} else if strict {
						return ""

					} else if remain {
						sb.WriteString("${")
						sb.WriteString(KtCvt.ToString(val))
						sb.WriteString("}")
					}

				} else {
					sb.WriteString(KtCvt.ToString(value))
				}
			}

			sIndex++
			index = KtStr.Index(exp, "${", sIndex)
		}

		exp = sb.String()
		if strings.Index(exp, "$$") >= 0 {
			exp = strings.ReplaceAll(exp, "$$", "$")
		}
	}

	return exp
}

type Read func(str string)

var splits = []byte("=:")

type BLinkedMap struct {
	b  int
	mp *Kt.LinkedMap
}

func ReadFunc(cfg Kt.Map, readMap *map[string]Read) Read {
	var bBuilder *strings.Builder
	var bAppend int
	var yB int
	var yMap *Kt.LinkedMap
	var ybMaps *Kt.Stack
	return func(str string) {
		sLen := len(str)
		sLast := sLen - 1
		for ; sLast >= 0; sLast-- {
			c := str[sLast]
			if c != '\r' && c != '\n' {
				break
			}
		}

		sLast++
		if sLast < sLen {
			str = str[0:sLast]
		}

		name := strings.TrimSpace(str)
		sLen = len(name)
		if sLen < 1 {
			return
		}

		chr := name[0]
		if bBuilder == nil {
			if chr == '#' || chr == ';' {
				return

			} else if chr == '{' && sLen == 2 && name[1] == '"' {
				bBuilder = &strings.Builder{}
				bAppend = 1
				return
			}

		} else if bAppend > 0 {
			if chr == '"' && sLen == 2 && name[1] == '}' {
				bAppend = 0

			} else {
				if bAppend > 1 {
					bBuilder.WriteString("\r\n")

				} else {
					bAppend = 2
				}

				bBuilder.WriteString(GetExp(cfg, str, false))
			}

			return
		}

		if sLen < 3 && !(sLen == 2 && name[1] == ':') {
			return
		}

		sLen = len(str)
		index := KtStr.IndexBytes(str, splits, 0)
		if index > 0 && index < sLen {
			if chr == '-' {
				chr = '+'
				if index > 1 {
					name = name[1:index]

				} else {
					name = "a"
				}

			} else {
				chr = str[index-1]
				if chr == '.' || chr == '#' || chr == ',' || chr == '+' || chr == '-' || chr == '$' {
					if index < 1 {
						return
					}

					name = str[0 : index-1]

				} else {
					chr = 0
					name = str[0:index]
				}
			}

			name = strings.TrimSpace(name)
			nLen := len(name)
			if nLen == 0 {
				return
			}

			// yml支持
			if str[index] == ':' {
				b := KtStr.IndentB(str)
				for {
					if b > yB || ybMaps == nil {
						break
					}

					if ybMaps.IsEmpty() {
						yB = 0
						yMap = nil
						break
					}

					ybMap := ybMaps.Pop().(*BLinkedMap)
					yB = ybMap.b
					yMap = ybMap.mp
				}

				// yaml字典key 例如 server:
				if len(strings.TrimSpace(str[index:])) == 1 {
					mp := new(Kt.LinkedMap).Init()
					var pMap Kt.Map
					if yMap == nil {
						pMap = cfg

					} else {
						pMap = yMap
					}

					if chr == '+' || chr == '-' {
						// 支持数组
						o := GetType(mp, name, nil, nil).(*list.List)
						if o == nil {
							o = list.New()
							pMap.Put(name, o)
						}

						o.PushBack(mp)

					} else {
						pMap.Put(name, mp)
					}

					if ybMaps == nil {
						ybMaps = new(Kt.Stack).Init()
					}

					if yMap != nil {
						ybMaps.Push(&BLinkedMap{b: yB, mp: yMap})
					}

					yB = b
					yMap = mp

					ybMaps.Push(&BLinkedMap{b: yB, mp: yMap})
					return
				}
			}

			eIndex := index
			index = strings.IndexByte(name, '|')
			if index > 0 {
				if nLen <= 1 {
					return
				}

				conds := KtStr.SplitByte(name, '|', true, index+1, 0)
				name = strings.TrimSpace(name[0:index])
				for _, cond := range conds {
					index = strings.IndexByte(cond, '&')
					if index > 0 {
						val := KtCvt.ToString(Get(cfg, cond[0:index]))
						if len(val) > 0 && KtStr.Match(cond[index+1:], false, val) {
							conds = nil
							break
						}

					} else if GetType(cfg, cond, KtCvt.Bool, false).(bool) {
						conds = nil
						break
					}
				}

				if conds != nil {
					return
				}
			}

			str = str[eIndex+1:]
			str = GetExp(cfg, KtStr.ToArg(str), false)
			if bBuilder != nil {
				if len(str) > 0 {
					bBuilder.WriteString("\r\n")
					bBuilder.WriteString(str)
				}

				str = bBuilder.String()
				bBuilder = nil
				bAppend = 0
			}

			if readMap != nil && name[0] == '@' {
				read := (*readMap)[name]
				if read != nil {
					read(str)
					return
				}
			}

			mp := Kt.If(yMap == nil, cfg, yMap).(Kt.Map)
			switch chr {
			case '.':
				o := mp.Get(name)
				mp.Put(name, Kt.If(o == nil, str, KtCvt.ToString(o)+str))
				break
			case '#':
				o := mp.Get(name)
				mp.Put(name, Kt.If(o == nil, str, KtCvt.ToString(o)+"\r\n"+str))
				break
			case ',':
				strs := KtStr.SplitStrBr(str, ",;", true, 0, false, 0, true).(*list.List)
				o := GetType(mp, name, nil, nil).(*list.List)
				if o == nil {
					o = strs
					mp.Put(name, o)

				} else {
					o.PushBackList(strs)
				}
				break
			case '+':
				o := GetType(mp, name, nil, nil).(*list.List)
				if o == nil {
					o = list.New()
					mp.Put(name, o)
				}

				o.PushBack(str)
				break
			case '-':
				mp.Remove(name)
				break
			case '$':
				// 配置复用
				mp.Put(name, Get(cfg, str))
			default:
				mp.Put(name, str)
				break
			}
		}
	}
}

func ReadIn(in *bufio.Reader, cfg Kt.Map, readMap *map[string]Read) Kt.Map {
	if in == nil {
		return cfg
	}

	if cfg == nil {
		cfg = Cfg{}
	}

	fun := ReadFunc(cfg, readMap)
	for {
		line, err := in.ReadString('\n')
		if line != "" {
			fun(line)
		}

		if err != nil {
			if err != io.EOF {
				Kt.Err(err, true)
			}

			break
		}
	}

	return cfg
}
