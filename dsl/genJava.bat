::// --out gen-go --out .\src\gen-go
cd  %~dp0/../
::thrift --gen java -out ../../axj2/axj-gw/src/main/java/ -I .\ -r .\gw.thrift
::protoc -I ./ ./dsl/gw.proto --plugin=protoc-gen-grpc-java=D:/tool/bin/protoc-gen-grpc-java.exe --grpc-java_out=="./dsl/gw.proto"
protoc -I ./ ./dsl/gw.proto --java_out="E:/open/axj/axj-gw/src/main/java/"
protoc --plugin=protoc-gen-grpc-java=C:/ProgramGreen/protoc-25.1-win64/bin/protoc-gen-grpc-java-1.60.1-windows-x86_64.exe --grpc-java_out="E:/open/axj/axj-gw/src/main/java/" --proto_path="./" "./dsl/gw.proto"