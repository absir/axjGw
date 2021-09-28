// --out gen-go --out .\src\gen-go
thrift --gen go -I .\dsl -r .\dsl\gw.thrift
thrift --gen go -I .\dsl -r .\dsl\gwInner.thrift