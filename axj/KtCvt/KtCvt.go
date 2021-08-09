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

func ToType(obj interface{}, typ reflect.Type) interface{} {
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
