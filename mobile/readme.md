# Mobile-SDK

gomobile sdk

go install golang.org/x/mobile/cmd/gomobile@latest

go install golang.org/x/mobile/cmd/gomobile@v0.0.0-20231127183840-76ac6878050a

gomobile init

set JAVA_TOOL_OPTIONS=-Dfile.encoding=utf-8
export JAVA_TOOL_OPTIONS=-Dfile.encoding=utf-8

## 生成命令

```shell
sh build.sh
```

## 测试环境

```
wss://gw.dev.yiyiny.com/gw
```

## 安卓环境依赖

- 1、下载 android studio
- 2、安装sdk

## 调用GRPC

1、 Java 实现

```java
@Bean
public class GRpcPassImpl extends PassGrpcServer {

}
```

2、GO增加配置

```yml
prods:
  store:
    1: ms-store.ms:8083
```

3、服务调用

> 假设 store 服务上存在路由: `/test/sendU` 即可如下调用

```go
client.Req("test/sendU", data, false, 30, func(s string, data []byte) {
	fmt.Println("send Back === " + s + KtUnsafe.BytesToString(data))
})
```

## 字典压缩

同级目录下增加 `uriDict.properties` 文件