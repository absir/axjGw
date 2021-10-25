// --out gen-go --out .\src\gen-go
// cd  %~dp0
thrift --gen go -I .\ -r .\dsl\gw.thrift
thrift --gen go -I .\ -r .\dsl\gwI.thrift