// --out gen-go --out .\src\gen-go
cd  %~dp0
thrift --gen java -out ../../axj2/axj-gw/src/main/java/ -I .\ -r .\gw.thrift