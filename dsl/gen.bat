// --out gen-go --out .\src\gen-go
cd  %~dp0
thrift --gen go -out ..\src\gen\ -I .\ -r .\gw.thrift
thrift --gen go -out ..\src\gen\ -I .\ -r .\gwI.thrift