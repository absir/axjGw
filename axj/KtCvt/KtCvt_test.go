package KtCvt

import (
	"reflect"
	"testing"
)

func TestCvt(t *testing.T) {
	type args struct {
		from interface{}
		to   reflect.Kind
	}

	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{"1=1", args{"1", reflect.Int}, 1},
		{"1.01=1", args{"1.01", reflect.Int}, 1},
		{"1.01=true", args{"1.01", reflect.Bool}, true},
		{"no=false", args{"no", reflect.Bool}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToType(tt.args.from, tt.args.to); got != tt.want {
				t.Errorf("ToType() = %v, want %v", got, tt.want)
			}
		})
	}
}
