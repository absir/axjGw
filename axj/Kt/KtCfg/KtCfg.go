package KtCfg

import (
	Kt2 "axj/Kt/Kt"
	KtCvt2 "axj/Kt/KtCvt"
	KtStr2 "axj/Kt/KtStr"
	"bufio"
	"container/list"
	"io"
	"os"
	"reflect"
	"strings"
)

// 配置字典别名
type Cfg map[interface{}]interface{}

func (c Cfg) Map() map[interface{}]interface{} {
	return c
}

func (c Cfg) Get(key interface{}) interface{} {
	return c[key]
}

func (c Cfg) Put(key interface{}, val interface{}) interface{} {
	_val := c[key]
	c[key] = val
	return _val
}

func (c Cfg) Remove(key interface{}) interface{} {
	_val := c[key]
	delete(c, key)
	return _val
}

// 配置获取
func Get(cfg Kt2.Map, name string) interface{} {
	val := Kt2.If(cfg == nil, nil, cfg.Get(name))
	if val == nil {
		if len(name) > 1 {
			chr := name[0]
			switch chr {
			case '%':
				return os.Getenv(name[1:])
				break
			case '@':
			case '$':
				return Kt2.If(cfg == nil, nil, cfg.Get(name[1:]))
				break
			}
		}

		if cfg != nil {
			i := -1
			for {
				i = KtStr2.LastIndex(name, ".", i)
				if i < 0 {
					break
				}

				c := cfg.Get(name[0:i])
				if c != nil {
					if mp, is := c.(Kt2.Map); is {
						return Get(mp, name[i+1:])
					}
				}
			}
		}
	}

	return val
}

func GetType(cfg Kt2.Map, name string, typ reflect.Type, dVal interface{}, tName string) interface{} {
	val := Get(cfg, name)
	if val != nil {
		if typ == nil {
			if _, is := val.(*list.List); !is {
				if reflect.TypeOf(val).Kind() != reflect.Array {
					if _, is := val.(*list.List); !is {
						val = []interface{}{val}
					}
				}

				ary := KtCvt2.ToType(Kt2.If(val == nil, dVal, val), typ).([]interface{})
				val = Kt2.ToList(ary)
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

	return KtCvt2.ToType(Kt2.If(val == nil, dVal, val), typ)
}

func GetExp(cfg Kt2.Map, exp string, strict bool) string {
	return GetExpRemain(cfg, exp, strict, false)
}

func GetExpRemain(cfg Kt2.Map, exp string, strict bool, remain bool) string {
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

			sIndex = KtStr2.Index(exp, "${", index)
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
					vals := KtStr2.SplitStr(val, "?:", true, 0)
					val := vals[0]
					if len(vals) == 3 {
						strict = false
						valD = Kt2.If(GetType(cfg, val.(string), KtCvt2.Bool, false, "").(bool), vals[1], vals[2])

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
						sb.WriteString(KtCvt2.ToString(val))
						sb.WriteString("}")
					}

				} else {
					sb.WriteString(KtCvt2.ToString(value))
				}
			}

			sIndex++
			index = KtStr2.Index(exp, "${", sIndex)
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

func ReadFunc(cfg Kt2.Map, readMap *map[string]Read) Read {
	var bBuilder *strings.Builder
	var bAppend int
	var yB int
	var yMap *Kt2.LinkedMap
	var yMaps *Kt2.Stack
	return func(str string) {
		str = strings.TrimSpace(str)
		length := len(str)
		if length < 1 {
			return
		}

		chr := str[0]
		if bBuilder == nil {
			if chr == '#' || chr == ';' {
				return

			} else if chr == '{' && length == 2 && str[1] == '"' {
				bBuilder = &strings.Builder{}
				bAppend = 1
				return
			}

		} else if bAppend > 0 {
			if chr == '"' && length == 2 && str[1] == '}' {
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

		if length < 3 {
			return
		}

		index := KtStr2.IndexBytes(str, splits, 0)
		if index > 0 && index < length {
			var name string
			chr = str[index-1]
			if chr == '.' || chr == '#' || chr == ',' || chr == '+' || chr == '-' {
				if index < 1 {
					return
				}

				name = str[0 : index-1]

			} else {
				chr = 0
				name = str[0:index]
			}

			name = strings.TrimSpace(name)
			length := len(name)
			if length == 0 {
				return
			}

			// yml支持
			if str[index] == ':' && len(strings.TrimSpace(str[index:])) == 1 {
				b := index - length
				if yB < b {
					if yMap != nil {
						if yMaps == nil {
							yMaps = new(Kt2.Stack).Init()
						}

						yMaps.Push(yMap)
					}

					lmap := new(Kt2.LinkedMap).Init()
					if yMap == nil {
						cfg.Put(name, lmap)

					} else {
						yMap.Put(name, lmap)
					}

					yMap = lmap

				} else {
					if b == 0 {
						// 根配置
						if yMaps != nil {
							yMaps.Clear()
						}

						yMap = new(Kt2.LinkedMap).Init()
						cfg.Put(name, yMap)

					} else {
						if yMaps == nil || yMaps.IsEmpty() {
							yMap = nil

						} else {
							yMap = yMaps.Pop().(*Kt2.LinkedMap)
						}
					}
				}

				yB = b
				return
			}

			eIndex := index
			index = strings.IndexByte(name, '|')
			if index > 0 {
				if length <= 1 {
					return
				}

				conds := KtStr2.SplitByte(name, '|', true, index+1, 0)
				name = strings.TrimSpace(name[0:index])
				for _, cond := range conds {
					index = strings.IndexByte(cond, '&')
					if index > 0 {
						val := KtCvt2.ToString(Get(cfg, cond[0:index]))
						if len(val) > 0 && KtStr2.Match(cond[index+1:], false, val) {
							conds = nil
							break
						}

					} else if GetType(cfg, cond, KtCvt2.Bool, false, "").(bool) {
						conds = nil
						break
					}
				}

				if conds != nil {
					return
				}
			}

			str = str[eIndex+1:]
			str = GetExp(cfg, KtStr2.ToArg(str), false)
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

			mp := Kt2.If(yMap == nil, cfg, yMap).(Kt2.Map)
			switch chr {
			case '.':
				o := mp.Get(name)
				mp.Put(name, Kt2.If(o == nil, str, KtCvt2.ToString(o)+str))
				break
			case '#':
				o := mp.Get(name)
				mp.Put(name, Kt2.If(o == nil, str, KtCvt2.ToString(o)+"\r\n"+str))
				break
			case ',':
				strs := KtStr2.SplitStrBr(str, ",;", true, 0, false, 0, true).(*list.List)
				o := GetType(mp, name, nil, nil, "").(*list.List)
				if o == nil {
					o = strs
					mp.Put(name, o)

				} else {
					o.PushBackList(strs)
				}
				break
			case '+':
				o := GetType(mp, name, nil, nil, "").(*list.List)
				if o == nil {
					o = list.New()
					mp.Put(name, o)
				}

				o.PushBack(str)
				break
			case '-':
				mp.Remove(name)
				break
			default:
				mp.Put(name, str)
				break
			}
		}
	}
}

func ReadIn(in *bufio.Reader, cfg Kt2.Map, readMap *map[string]Read) Kt2.Map {
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
				Kt2.Err(err, true)
			}

			break
		}
	}

	return cfg
}
