package KtCvt

import (
	"axj/Kt"
	"container/list"
	"fmt"
	"reflect"
	"strconv"
)

var Bool = reflect.TypeOf(*new(bool))
var String = reflect.TypeOf(*new(string))
var Int = reflect.TypeOf(*new(int))
var Int8 = reflect.TypeOf(*new(int8))
var Int16 = reflect.TypeOf(*new(int16))
var Int32 = reflect.TypeOf(*new(int32))
var Int64 = reflect.TypeOf(*new(int64))
var UInt = reflect.TypeOf(*new(uint))
var UInt8 = reflect.TypeOf(*new(uint8))
var UInt16 = reflect.TypeOf(*new(uint16))
var UInt32 = reflect.TypeOf(*new(uint32))
var UInt64 = reflect.TypeOf(*new(uint64))
var Float32 = reflect.TypeOf(*new(float32))
var Float64 = reflect.TypeOf(*new(float64))
var Complex64 = reflect.TypeOf(*new(complex64))
var Complex128 = reflect.TypeOf(*new(complex128))
var Array = reflect.TypeOf(*new([]interface{}))

func Safe(obj interface{}) interface{} {
	if mp, is := obj.(map[interface{}]interface{}); is {
		nmp := map[string]interface{}{}
		for key, val := range mp {
			nmp[ToString(key)] = Safe(val)
		}

		return nmp
	}

	if list, is := obj.(*list.List); is {
		array := make([]interface{}, list.Len())
		i := 0
		for el := list.Front(); el != nil; el = el.Next() {
			array[i] = Safe(el.Value)
			i++
		}

		return array
	}

	return obj
}

func ToType(obj interface{}, typ reflect.Type) interface{} {
	if typ == nil {
		return obj
	}

	switch typ.Kind() {
	case reflect.Bool:
		return ToBool(obj)
	case reflect.String:
		return ToString(obj)
	case reflect.Int:
		return int(ToInt64(obj))
	case reflect.Int8:
		return int8(ToInt32(obj))
	case reflect.Int16:
		return int16(ToInt32(obj))
	case reflect.Int32:
		return ToInt32(obj)
	case reflect.Int64:
		return ToInt64(obj)
	case reflect.Uint:
		return uint(ToUInt64(obj))
	case reflect.Uint8:
		return uint8(ToUInt64(obj))
	case reflect.Uint16:
		return uint16(ToUInt64(obj))
	case reflect.Uint32:
		return uint32(ToUInt64(obj))
	case reflect.Uint64:
		return ToUInt64(obj)
	case reflect.Float32:
		return ToFloat32(obj)
	case reflect.Float64:
		return ToFloat32(obj)
	case reflect.Complex64:
		return ToComplex64(obj)
	case reflect.Complex128:
		return ToComplex128(obj)
	case reflect.Interface:
		return obj
	default:
		break
	}

	if obj == nil {
		return nil
	}

	oTyp := reflect.TypeOf(obj)
	if oTyp == typ {
		return obj
	}

	if oTyp.ConvertibleTo(typ) {
		return obj
	}

	if typ.Kind() == reflect.Array && typ.Elem().Kind() == reflect.Interface {
		switch obj.(type) {
		case *list.List:
			return Kt.ToArray(obj.(*list.List))
			break
		}
	}

	return nil
}

func ToBool(obj interface{}) bool {
	if obj != nil {
		typ := reflect.TypeOf(obj)
		switch typ.Kind() {
		case reflect.Bool:
			return obj.(bool)
		case reflect.String:
			str := obj.(string)
			if len(str) > 0 {
				chr := str[0]
				return !(chr == 0 || chr == '0' || chr == 'F' || chr == 'N' || chr == 'f' || chr == 'n')
			}

			return false
		case reflect.Int:
			return obj.(int) != 0
		case reflect.Int8:
			return obj.(int8) != 0
		case reflect.Int16:
			return obj.(int16) != 0
		case reflect.Int32:
			return obj.(int32) != 0
		case reflect.Int64:
			return obj.(int64) != 0
		case reflect.Uint:
			return obj.(uint) != 0
		case reflect.Uint8:
			return obj.(uint8) != 0
		case reflect.Uint16:
			return obj.(uint16) != 0
		case reflect.Uint32:
			return obj.(uint32) != 0
		case reflect.Uint64:
			return obj.(uint64) != 0
		case reflect.Float32:
			return obj.(float32) != 0
		case reflect.Float64:
			return obj.(float64) != 0
		case reflect.Complex64:
			return obj.(complex64) != 0
		case reflect.Complex128:
			return obj.(complex128) != 0
		default:
			return true
		}
	}

	return false
}

func ToString(obj interface{}) string {
	if obj != nil {
		typ := reflect.TypeOf(obj)
		switch typ.Kind() {
		case reflect.Bool:
			if obj.(bool) {
				return "1"
			}

			return "0"
		case reflect.String:
			return obj.(string)
		case reflect.Int:
			return strconv.Itoa(obj.(int))
		case reflect.Int8:
			return strconv.Itoa(int(obj.(int8)))
		case reflect.Int16:
			return strconv.Itoa(int(obj.(int16)))
		case reflect.Int32:
			return strconv.Itoa(int(obj.(int32)))
		case reflect.Int64:
			return strconv.FormatInt(obj.(int64), 10)
		case reflect.Uint:
			return strconv.FormatUint(uint64(obj.(uint)), 10)
		case reflect.Uint8:
			return strconv.FormatUint(uint64(obj.(uint8)), 10)
		case reflect.Uint16:
			return strconv.FormatUint(uint64(obj.(uint16)), 10)
		case reflect.Uint32:
			return strconv.FormatUint(uint64(obj.(uint32)), 10)
		case reflect.Uint64:
			return strconv.FormatUint(obj.(uint64), 10)
		case reflect.Float32:
			return strconv.FormatFloat(float64(obj.(float32)), 'e', -1, 32)
		case reflect.Float64:
			return strconv.FormatFloat(obj.(float64), 'e', -1, 64)
		case reflect.Complex64:
			return strconv.FormatComplex(complex128(obj.(complex64)), 'e', 0, 64)
		case reflect.Complex128:
			return strconv.FormatComplex(obj.(complex128), 'e', 0, 128)
		default:
			return fmt.Sprintf("%s&%p", typ.Name(), &obj)
		}
	}

	return ""
}

func ToInt32(obj interface{}) int32 {
	if obj != nil {
		typ := reflect.TypeOf(obj)
		switch typ.Kind() {
		case reflect.Bool:
			if obj.(bool) {
				return 1
			}

			return 0
		case reflect.String:
			i, err := strconv.Atoi(obj.(string))
			if err != nil {
				f, err := strconv.ParseFloat(obj.(string), 10)
				if err == nil {
					return int32(f)
				}

				i = 0
			}

			return int32(i)
		case reflect.Int:
			return int32(obj.(int))
		case reflect.Int8:
			return int32(obj.(int8))
		case reflect.Int16:
			return int32(obj.(int16))
		case reflect.Int32:
			return obj.(int32)
		case reflect.Int64:
			return int32(obj.(int64))
		case reflect.Uint:
			return int32(obj.(uint))
		case reflect.Uint8:
			return int32(obj.(uint8))
		case reflect.Uint16:
			return int32(obj.(uint16))
		case reflect.Uint32:
			return int32(obj.(uint32))
		case reflect.Uint64:
			return int32(obj.(uint64))
		case reflect.Float32:
			return int32(obj.(float32))
		case reflect.Float64:
			return int32(obj.(float64))
		case reflect.Complex64:
			return int32(real(obj.(complex64)))
		case reflect.Complex128:
			return int32(real(obj.(complex128)))
		default:
			return 0
		}
	}

	return 0
}

func ToInt64(obj interface{}) int64 {
	if obj != nil {
		typ := reflect.TypeOf(obj)
		switch typ.Kind() {
		case reflect.Bool:
			if obj.(bool) {
				return 1
			}

			return 0
		case reflect.String:
			i, err := strconv.Atoi(obj.(string))
			if err != nil {
				f, err := strconv.ParseFloat(obj.(string), 10)
				if err == nil {
					return int64(f)
				}

				i = 0
			}

			return int64(i)
		case reflect.Int:
			return int64(obj.(int))
		case reflect.Int8:
			return int64(obj.(int8))
		case reflect.Int16:
			return int64(obj.(int16))
		case reflect.Int32:
			return int64(obj.(int32))
		case reflect.Int64:
			return obj.(int64)
		case reflect.Uint:
			return int64(obj.(uint))
		case reflect.Uint8:
			return int64(obj.(uint8))
		case reflect.Uint16:
			return int64(obj.(uint16))
		case reflect.Uint32:
			return int64(obj.(uint32))
		case reflect.Uint64:
			return int64(obj.(uint64))
		case reflect.Float32:
			return int64(obj.(float32))
		case reflect.Float64:
			return int64(obj.(float64))
		case reflect.Complex64:
			return int64(real(obj.(complex64)))
		case reflect.Complex128:
			return int64(real(obj.(complex128)))
		default:
			return 0
		}
	}

	return 0
}

func ToUInt64(obj interface{}) uint64 {
	if obj != nil {
		typ := reflect.TypeOf(obj)
		switch typ.Kind() {
		case reflect.Bool:
			if obj.(bool) {
				return 1
			}

			return 0
		case reflect.String:
			i, err := strconv.Atoi(obj.(string))
			if err != nil {
				f, err := strconv.ParseFloat(obj.(string), 10)
				if err == nil {
					return uint64(f)
				}

				i = 0
			}

			return uint64(i)
		case reflect.Int:
			return uint64(obj.(int))
		case reflect.Int8:
			return uint64(obj.(int8))
		case reflect.Int16:
			return uint64(obj.(int16))
		case reflect.Int32:
			return uint64(obj.(int32))
		case reflect.Int64:
			return uint64(obj.(int64))
		case reflect.Uint:
			return uint64(obj.(uint))
		case reflect.Uint8:
			return uint64(obj.(uint8))
		case reflect.Uint16:
			return uint64(obj.(uint16))
		case reflect.Uint32:
			return uint64(obj.(uint32))
		case reflect.Uint64:
			return obj.(uint64)
		case reflect.Float32:
			return uint64(obj.(float32))
		case reflect.Float64:
			return uint64(obj.(float64))
		case reflect.Complex64:
			return uint64(real(obj.(complex64)))
		case reflect.Complex128:
			return uint64(real(obj.(complex128)))
		default:
			return 0
		}
	}

	return 0
}

func ToFloat32(obj interface{}) float32 {
	if obj != nil {
		typ := reflect.TypeOf(obj)
		switch typ.Kind() {
		case reflect.Bool:
			if obj.(bool) {
				return 1
			}

			return 0
		case reflect.String:
			f, err := strconv.ParseFloat(obj.(string), 10)
			if err == nil {
				return float32(f)
			}

			return 0
		case reflect.Int:
			return float32(obj.(int))
		case reflect.Int8:
			return float32(obj.(int8))
		case reflect.Int16:
			return float32(obj.(int16))
		case reflect.Int32:
			return float32(obj.(int32))
		case reflect.Int64:
			return float32(obj.(int64))
		case reflect.Uint:
			return float32(obj.(uint))
		case reflect.Uint8:
			return float32(obj.(uint8))
		case reflect.Uint16:
			return float32(obj.(uint16))
		case reflect.Uint32:
			return float32(obj.(uint32))
		case reflect.Uint64:
			return float32(obj.(uint64))
		case reflect.Float32:
			return obj.(float32)
		case reflect.Float64:
			return float32(obj.(float64))
		case reflect.Complex64:
			return float32(real(obj.(complex64)))
		case reflect.Complex128:
			return float32(real(obj.(complex128)))
		default:
			return 0
		}
	}

	return 0
}

func ToFloat64(obj interface{}) float64 {
	if obj != nil {
		typ := reflect.TypeOf(obj)
		switch typ.Kind() {
		case reflect.Bool:
			if obj.(bool) {
				return 1
			}

			return 0
		case reflect.String:
			f, err := strconv.ParseFloat(obj.(string), 10)
			if err == nil {
				return float64(f)
			}

			return 0
		case reflect.Int:
			return float64(obj.(int))
		case reflect.Int8:
			return float64(obj.(int8))
		case reflect.Int16:
			return float64(obj.(int16))
		case reflect.Int32:
			return float64(obj.(int32))
		case reflect.Int64:
			return float64(obj.(int64))
		case reflect.Uint:
			return float64(obj.(uint))
		case reflect.Uint8:
			return float64(obj.(uint8))
		case reflect.Uint16:
			return float64(obj.(uint16))
		case reflect.Uint32:
			return float64(obj.(uint32))
		case reflect.Uint64:
			return float64(obj.(uint64))
		case reflect.Float32:
			return float64(obj.(float32))
		case reflect.Float64:
			return obj.(float64)
		case reflect.Complex64:
			return float64(real(obj.(complex64)))
		case reflect.Complex128:
			return float64(real(obj.(complex128)))
		default:
			return 0
		}
	}

	return 0
}

func Complex64F(real float64) complex64 {
	return complex64(complex(real, 0))
}

func Complex128F(real float64) complex128 {
	return complex(real, 0)
}

func ToComplex64(obj interface{}) complex64 {
	if obj != nil {
		typ := reflect.TypeOf(obj)
		switch typ.Kind() {
		case reflect.Bool:
			if obj.(bool) {
				return 1
			}

			return 0
		case reflect.String:
			c, err := strconv.ParseComplex(obj.(string), 10)
			if err != nil {
				f, err := strconv.ParseFloat(obj.(string), 10)
				if err == nil {
					return Complex64F(f)
				}

				c = complex(0, 0)
			}

			return complex64(c)
		case reflect.Int:
			return Complex64F(float64(obj.(int)))
		case reflect.Int8:
			return Complex64F(float64(obj.(int8)))
		case reflect.Int16:
			return Complex64F(float64(obj.(int16)))
		case reflect.Int32:
			return Complex64F(float64(obj.(int32)))
		case reflect.Int64:
			return Complex64F(float64(obj.(int64)))
		case reflect.Uint:
			return Complex64F(float64(obj.(uint)))
		case reflect.Uint8:
			return Complex64F(float64(obj.(uint8)))
		case reflect.Uint16:
			return Complex64F(float64(obj.(uint16)))
		case reflect.Uint32:
			return Complex64F(float64(obj.(uint32)))
		case reflect.Uint64:
			return Complex64F(float64(obj.(uint64)))
		case reflect.Float32:
			return Complex64F(float64(obj.(float32)))
		case reflect.Float64:
			return Complex64F(obj.(float64))
		case reflect.Complex64:
			return obj.(complex64)
		case reflect.Complex128:
			return complex64(obj.(complex128))
		default:
			return 0
		}
	}

	return 0
}

func ToComplex128(obj interface{}) complex128 {
	if obj != nil {
		typ := reflect.TypeOf(obj)
		switch typ.Kind() {
		case reflect.Bool:
			if obj.(bool) {
				return 1
			}

			return 0
		case reflect.String:
			c, err := strconv.ParseComplex(obj.(string), 10)
			if err != nil {
				f, err := strconv.ParseFloat(obj.(string), 10)
				if err == nil {
					return Complex128F(f)
				}

				c = complex(0, 0)
			}

			return c
		case reflect.Int:
			return Complex128F(float64(obj.(int)))
		case reflect.Int8:
			return Complex128F(float64(obj.(int8)))
		case reflect.Int16:
			return Complex128F(float64(obj.(int16)))
		case reflect.Int32:
			return Complex128F(float64(obj.(int32)))
		case reflect.Int64:
			return Complex128F(float64(obj.(int64)))
		case reflect.Uint:
			return Complex128F(float64(obj.(uint)))
		case reflect.Uint8:
			return Complex128F(float64(obj.(uint8)))
		case reflect.Uint16:
			return Complex128F(float64(obj.(uint16)))
		case reflect.Uint32:
			return Complex128F(float64(obj.(uint32)))
		case reflect.Uint64:
			return Complex128F(float64(obj.(uint64)))
		case reflect.Float32:
			return Complex128F(float64(obj.(float32)))
		case reflect.Float64:
			return Complex128F(obj.(float64))
		case reflect.Complex64:
			return complex128(obj.(complex64))
		case reflect.Complex128:
			return obj.(complex128)
		default:
			return 0
		}
	}

	return 0
}

type ArrayIs struct {
	New  func(size int) interface{}
	Size func(array interface{}) int
	Get  func(array interface{}, i int) interface{}
	Set  func(array interface{}, i int, el interface{})
}

var BoolIs = ArrayIs{func(size int) interface{} {
	return make([]bool, size)
}, func(array interface{}) int {
	return len(array.([]bool))
}, func(array interface{}, i int) interface{} {
	return array.([]bool)[i]
}, func(array interface{}, i int, el interface{}) {
	array.([]bool)[i] = el.(bool)
}}

var StringIs = ArrayIs{func(size int) interface{} {
	return make([]string, size)
}, func(array interface{}) int {
	return len(array.([]string))
}, func(array interface{}, i int) interface{} {
	return array.([]string)[i]
}, func(array interface{}, i int, el interface{}) {
	array.([]string)[i] = el.(string)
}}

var IntIs = ArrayIs{func(size int) interface{} {
	return make([]int, size)
}, func(array interface{}) int {
	return len(array.([]int))
}, func(array interface{}, i int) interface{} {
	return array.([]int)[i]
}, func(array interface{}, i int, el interface{}) {
	array.([]int)[i] = el.(int)
}}

var Int8Is = ArrayIs{func(size int) interface{} {
	return make([]int8, size)
}, func(array interface{}) int {
	return len(array.([]int8))
}, func(array interface{}, i int) interface{} {
	return array.([]int8)[i]
}, func(array interface{}, i int, el interface{}) {
	array.([]int8)[i] = el.(int8)
}}

var Int16Is = ArrayIs{func(size int) interface{} {
	return make([]int16, size)
}, func(array interface{}) int {
	return len(array.([]int16))
}, func(array interface{}, i int) interface{} {
	return array.([]int16)[i]
}, func(array interface{}, i int, el interface{}) {
	array.([]int16)[i] = el.(int16)
}}

var Int32Is = ArrayIs{func(size int) interface{} {
	return make([]int32, size)
}, func(array interface{}) int {
	return len(array.([]int32))
}, func(array interface{}, i int) interface{} {
	return array.([]int32)[i]
}, func(array interface{}, i int, el interface{}) {
	array.([]int32)[i] = el.(int32)
}}

var Int64Is = ArrayIs{func(size int) interface{} {
	return make([]int64, size)
}, func(array interface{}) int {
	return len(array.([]int64))
}, func(array interface{}, i int) interface{} {
	return array.([]int64)[i]
}, func(array interface{}, i int, el interface{}) {
	array.([]int64)[i] = el.(int64)
}}

var UInt8Is = ArrayIs{func(size int) interface{} {
	return make([]uint8, size)
}, func(array interface{}) int {
	return len(array.([]uint8))
}, func(array interface{}, i int) interface{} {
	return array.([]uint8)[i]
}, func(array interface{}, i int, el interface{}) {
	array.([]uint8)[i] = el.(uint8)
}}

var UInt16Is = ArrayIs{func(size int) interface{} {
	return make([]uint16, size)
}, func(array interface{}) int {
	return len(array.([]uint16))
}, func(array interface{}, i int) interface{} {
	return array.([]uint16)[i]
}, func(array interface{}, i int, el interface{}) {
	array.([]uint16)[i] = el.(uint16)
}}

var UInt32Is = ArrayIs{func(size int) interface{} {
	return make([]uint32, size)
}, func(array interface{}) int {
	return len(array.([]uint32))
}, func(array interface{}, i int) interface{} {
	return array.([]uint32)[i]
}, func(array interface{}, i int, el interface{}) {
	array.([]uint32)[i] = el.(uint32)
}}

var UInt64Is = ArrayIs{func(size int) interface{} {
	return make([]uint64, size)
}, func(array interface{}) int {
	return len(array.([]uint64))
}, func(array interface{}, i int) interface{} {
	return array.([]uint64)[i]
}, func(array interface{}, i int, el interface{}) {
	array.([]uint64)[i] = el.(uint64)
}}

var Float32Is = ArrayIs{func(size int) interface{} {
	return make([]float32, size)
}, func(array interface{}) int {
	return len(array.([]float32))
}, func(array interface{}, i int) interface{} {
	return array.([]float32)[i]
}, func(array interface{}, i int, el interface{}) {
	array.([]float32)[i] = el.(float32)
}}

var Float64Is = ArrayIs{func(size int) interface{} {
	return make([]float64, size)
}, func(array interface{}) int {
	return len(array.([]float64))
}, func(array interface{}, i int) interface{} {
	return array.([]float64)[i]
}, func(array interface{}, i int, el interface{}) {
	array.([]float64)[i] = el.(float64)
}}

var Complex64Is = ArrayIs{func(size int) interface{} {
	return make([]complex64, size)
}, func(array interface{}) int {
	return len(array.([]complex64))
}, func(array interface{}, i int) interface{} {
	return array.([]complex64)[i]
}, func(array interface{}, i int, el interface{}) {
	array.([]complex64)[i] = el.(complex64)
}}

var Complex128Is = ArrayIs{func(size int) interface{} {
	return make([]complex128, size)
}, func(array interface{}) int {
	return len(array.([]complex128))
}, func(array interface{}, i int) interface{} {
	return array.([]complex128)[i]
}, func(array interface{}, i int, el interface{}) {
	array.([]complex128)[i] = el.(complex128)
}}

var InterfaceIs = ArrayIs{func(size int) interface{} {
	return make([]interface{}, size)
}, func(array interface{}) int {
	return len(array.([]interface{}))
}, func(array interface{}, i int) interface{} {
	return array.([]interface{})[i]
}, func(array interface{}, i int, el interface{}) {
	array.([]interface{})[i] = el.(interface{})
}}

func ForArrayIs(typ reflect.Type) *ArrayIs {
	if typ == nil {
		return &InterfaceIs
	}

	switch typ.Kind() {
	case reflect.Bool:
		return &BoolIs
	case reflect.String:
		return &StringIs
	case reflect.Int:
		return &IntIs
	case reflect.Int8:
		return &Int8Is
	case reflect.Int16:
		return &Int16Is
	case reflect.Int32:
		return &Int32Is
	case reflect.Int64:
		return &Int64Is
	case reflect.Uint8:
		return &UInt8Is
	case reflect.Uint16:
		return &UInt16Is
	case reflect.Uint32:
		return &UInt32Is
	case reflect.Uint64:
		return &UInt64Is
	case reflect.Float32:
		return &Float32Is
	case reflect.Float64:
		return &Float64Is
	case reflect.Complex64:
		return &Complex64Is
	case reflect.Complex128:
		return &Complex128Is
	case reflect.Interface:
		return &InterfaceIs
	}

	return nil
}

func ToArray(lst *list.List, typ reflect.Type) interface{} {
	if lst == nil {
		return nil
	}

	is := ForArrayIs(typ)
	if is == nil {
		return nil
	}

	array := is.New(lst.Len())
	i := 0
	for el := lst.Front(); el != nil; el = el.Next() {
		is.Set(array, i, ToType(el.Value, typ))
		i++
	}

	return array
}
