module axjGW

go 1.16

require (
	axj v0.0.0
	github.com/apache/thrift v0.14.2
	gw v0.0.0
)

replace (
	axj v0.0.0 => ./../axj
	gw v0.0.0 => ./../gen-go/gw
)
