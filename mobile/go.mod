module mobile

go 1.19

require (
	axj v0.0.0
	gw v0.0.0
)

require golang.org/x/mobile v0.0.0-20211207041440-4e6c2922fdee // indirect

replace axj v0.0.0 => ./../axj

replace gw v0.0.0 => ./../src
