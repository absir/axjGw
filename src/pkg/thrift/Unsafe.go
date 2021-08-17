package thrift

import (
	"context"
	"github.com/apache/thrift/lib/go/thrift"
	"reflect"
)

func GetHeaderProtocol(context context.Context) *thrift.THeaderProtocol {
	helper, _ := thrift.GetResponseHelper(context)
	return reflect.ValueOf(helper.THeaderResponseHelper).Elem().FieldByName("proto").Interface().(*thrift.THeaderProtocol)
}
