package KtCvt

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCvt(t *testing.T) {

	ptr := ToType(1, reflect.TypeOf(new(int))).(*int)
	fmt.Println(*ptr)

	type args struct {
		from interface{}
		to   reflect.Type
	}

	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{"1=1", args{"1", Int}, 1},
		{"1.01=1", args{"1.01", Int}, 1},
		{"1.01=true", args{"1.01", Bool}, true},
		{"no=false", args{"no", Bool}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToType(tt.args.from, tt.args.to); got != tt.want {
				t.Errorf("ToType() = %v, want %v", got, tt.want)
			}
		})
	}
}
