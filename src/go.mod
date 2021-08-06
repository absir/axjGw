module axjGW

go 1.16

require (
	axj v0.0.0
	gw v0.0.0
)

replace (
	gw v0.0.0 => ./../gen-go/gw
	axj v0.0.0 => ./../axj
)