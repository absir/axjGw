package KtStr

import (
	KtCvt2 "axj/Kt/KtCvt"
	KtJson2 "axj/Kt/KtJson"
	"testing"
)

func TestCompareV(t *testing.T) {
	type args struct {
		from string
		to   string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"1,2", args{"1", "2"}, -1},
		{"0a,1", args{"0a", "1"}, 1},
		{"1.0,1.0", args{"1.0", "1.0"}, 0},
		{"0.1,0.0.1", args{"0.1", "0.0.1"}, 1},
		{"0.0,0.0.1", args{"0.0", "0.0.1"}, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompareV(tt.args.from, tt.args.to); got != tt.want {
				t.Errorf("CompareV() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIndex(t *testing.T) {
	type args struct {
		s      string
		substr string
		from   int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"1", args{"1232323", "23", 1}, 1},
		{"3", args{"1232323", "23", 3}, 3},
		{"4", args{"1232323", "23", 4}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Index(tt.args.s, tt.args.substr, tt.args.from); got != tt.want {
				t.Errorf("Index() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLastIndex(t *testing.T) {
	type args struct {
		s      string
		substr string
		from   int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"1", args{"1232323", "23", -1}, 5},
		{"3", args{"1232323", "23", 4}, 3},
		{"4", args{"1232323", "23", 2}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LastIndex(tt.args.s, tt.args.substr, tt.args.from); got != tt.want {
				t.Errorf("LastIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSim(t *testing.T) {
	type args struct {
		str string
		to  string
	}
	tests := []struct {
		name string
		args args
		want float32
	}{
		{"abc", args{"abc", "add"}, 0.3333333},
		{"123", args{"abcdefg", "abcdefg"}, 1},
		{"123", args{"abcddfg", "abcdefg"}, 0.85714287},
		{"123", args{"abdddfg", "abcdefg"}, 0.71428573},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret := Sim(tt.args.str, tt.args.to)
			if ret != tt.want {
				t.Errorf("SplitStrBr() = %v, want %v", ret, tt.want)
			}
		})
	}
}

func TestCap(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"abc", args{"abc"}, "Abc"},
		{"123", args{"123"}, "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret := Cap(tt.args.str)
			if ret != tt.want {
				t.Errorf("SplitStrBr() = %v, want %v", ret, tt.want)
			}
		})
	}
}

func TestUnCap(t *testing.T) {
	type args struct {
		str    string
		strict bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"abc", args{"ABC", true}, "ABC"},
		{"abc", args{"ABC", false}, "aBC"},
		{"123", args{"AbC", true}, "abC"},
		{"123", args{"AbC", false}, "abC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret := UnCap(tt.args.str, tt.args.strict)
			if ret != tt.want {
				t.Errorf("SplitStrBr() = %v, want %v", ret, tt.want)
			}
		})
	}
}

func TestSplit(t *testing.T) {
	type args struct {
		s     string
		sps   string
		trim  bool
		start int
		br    bool
		typ   int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"1", args{" 1,2, 3 ", ",", true, 0, false, 0}, "[\"1\",\"2\",\"3\"]"},
		{"1", args{" 1,,,2+ 3 ", ",=+", true, 0, false, 0}, "[\"1\",\"\",\"\",\"2\",\"3\"]"},
		{"1", args{" 1,{1,[2=3+++4]} ", ",=+", true, 0, true, 0}, "[\"1\",{\"1\":[\"2\",\"3\",\"\",\"\",\"4\"]}]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret := SplitStrBr(tt.args.s, tt.args.sps, tt.args.trim, tt.args.start, tt.args.br, tt.args.typ, false)
			json, err := KtJson2.ToJsonStr(KtCvt2.Safe(ret))
			if err != nil {
				t.Error(err)
			}

			if json != tt.want {
				t.Errorf("SplitStrBr() = %v, want %v", json, tt.want)
			}
		})
	}
}
