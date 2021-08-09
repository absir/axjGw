package KtCfg

import (
	"axj/Kt"
	"axj/KtCvt"
	"axj/KtStr"
	"os"
	"reflect"
	"strings"
)

// 配置字典别名
type Cfg map[string]interface{}

// 配置获取
func Get(cfg Cfg, name string) interface{} {
	val := Kt.If(cfg == nil, nil, cfg[name])
	if val == nil {
		if len(name) > 1 {
			chr := name[0]
			switch chr {
			case '%':
				return os.Getenv(name[1:])
				break
			case '@':
			case '$':
				return Kt.If(cfg == nil, nil, cfg[name[1:]])
				break
			}
		}
	}

	return val
}

func GetType(cfg Cfg, name string, typ reflect.Type, dVal interface{}, tName string) interface{} {
	val := Get(cfg, name)
	if val != nil {
		switch typ.Kind() {
		case reflect.Array:
			if reflect.TypeOf(val).Kind() != reflect.Array {
				val = [1]interface{}{val}
				cfg[name] = val
			}
			break
		default:
			break
		}
	}

	return KtCvt.ToType(Kt.If(val == nil, dVal, val), typ)
}

func GetExp(cfg Cfg, exp string, strict bool) string {
	return GetExpRemain(cfg, exp, strict, false)
}

func GetExpRemain(cfg Cfg, exp string, strict bool, remain bool) string {
	index := strings.Index(exp, "${")
	length := len(exp)
	if index >= 0 && index < length-2 {
		sb := new(strings.Builder)
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
						valD = Kt.If(GetType(cfg, val.(string), KtCvt.Bool, false, "").(bool), vals[1], vals[2])

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

func readFunc(cfg Cfg, read Read) Read {
	var bBuilder *strings.Builder
	var bAppend int
	var yB int
	var yMap *Kt.LinkedMap
	var yMaps *Kt.Stack
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
				bBuilder = new(strings.Builder)
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

		index := KtStr.IndexBytes(str, splits, 0)
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
							yMaps = new(Kt.Stack)
						}

						yMaps.Push(yMap)
					}

					lmap := new(Kt.LinkedMap)
					if yMap == nil {
						cfg[name] = lmap

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

						yMap = new(Kt.LinkedMap)
						cfg[name] = yMap

					} else {
						if yMaps == nil || yMaps.IsEmpty() {
							yMap = nil

						} else {
							yMap = yMaps.Pop().(*Kt.LinkedMap)
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

				conds := KtStr.SplitStr(name, "|", true, index+1)
				name = name.substring(0, index).trim()
				for String
			cond:
				conds) {
					index = cond.indexOf('&')
					if index > 0 {
						String
						val = KtCvt.to(getVal(cond.substring(0, index), cfgMap), String.class, null)
						if val != null && KtStr.match(cond.substring(index+1), false, val) {
							conds = null
							break
						}

					} else if getValClass(cond, cfgMap, boolean.class, false, null) {
						conds = null
						break
					}
				}

				if conds != null {
					return null
				}
			}

			s = s.substring(eIndex + 1)
			s = getExp(KtStr.toArg(s), false, cfgMap)
			if bBuilder != null {
				if s.length() > 0 {
					bBuilder.append("\r\n")
					bBuilder.append(s)
				}

				s = bBuilder.toString()
				bBuilder = null
				bAppend = 0
			}

			if funcMap != null && name.charAt(0) == '@' {
				KtB.Func1 < Void, String > func1 = funcMap.get(name)
				if func1 != null {
					func1.do1(s)
					return null
				}
			}

			Map
			map = yMap == null ? cfgMap:
			yMap
			Object
			o
			switch chr {
			case '.':
				o = map.get(name)
				map.put(name, o == null ? s:
				(o + s))
				break
			case '#':
				o = map.get(name)
				map.put(name, o == null ? s:
				(o + "\r\n" + s))
				break
			case ',':
				o = getValClass(name, map, List.class, null, null)
				if o == null {
					o = new
					ArrayList < Object > ()
					map.put(name, o)
				}
				(List < Object >)
				o).addAll(KtStr.splitStr(s, ",;", true, 0))
				break
			case '+':
				o = getValClass(name, map, List.class, null, null)
				if o == null {
					o = new
					ArrayList < Object > ()
					map.put(name, o)
				}
				(List < Object >)
				o).add(s)
				break
			case '-':
				map.remove(name)
				break
			default:
				map.put(name, s)
				break
			}
		}
	}
}
