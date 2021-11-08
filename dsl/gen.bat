:: --out gen-go --out .\src\gen-go
cd  %~dp0/../
::thrift --gen go -out ..\src\gen\ -I .\ -r .\gw.thrift
::thrift --gen go -out ..\src\gen\ -I .\ -r .\gwI.thrift
::protoc -I ./ ./gw.proto --go_out=plugins=grpc:../src/gen/
::cd ../
protoc -I ./ ./dsl/gw.proto --go_out=plugins=grpc:./src/gen/
protoc -I ./ ./dsl/gwI.proto --go_out=plugins=grpc:./src/gen/