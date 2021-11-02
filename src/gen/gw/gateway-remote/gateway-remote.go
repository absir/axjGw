// Code generated by Thrift Compiler (0.15.0). DO NOT EDIT.

package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	thrift "github.com/apache/thrift/lib/go/thrift"
	"gw"
)

var _ = gw.GoUnusedProtection__

func Usage() {
  fmt.Fprintln(os.Stderr, "Usage of ", os.Args[0], " [-h host:port] [-u url] [-f[ramed]] function [arg1 [arg2...]]:")
  flag.PrintDefaults()
  fmt.Fprintln(os.Stderr, "\nFunctions:")
  fmt.Fprintln(os.Stderr, "  bool close(i64 cid, string reason)")
  fmt.Fprintln(os.Stderr, "  bool kick(i64 cid, string bytes)")
  fmt.Fprintln(os.Stderr, "  bool rid(i64 cid, string name, i32 rid)")
  fmt.Fprintln(os.Stderr, "  bool rids(i64 cid,  rids)")
  fmt.Fprintln(os.Stderr, "  bool push(i64 cid, string uri, string bytes)")
  fmt.Fprintln(os.Stderr, "  bool gPush(string gid, string uri, string bytes, i32 qs, string unique, bool queue)")
  fmt.Fprintln(os.Stderr, "  bool gConn(i64 cid, string gid, string unique)")
  fmt.Fprintln(os.Stderr, "  bool gDisc(i64 cid, string gid, string unique, i32 connVer)")
  fmt.Fprintln(os.Stderr, "  bool gLasts(string gid, i64 cid, string unique, i64 lastId)")
  fmt.Fprintln(os.Stderr, "  bool send(string fromId, string toId, string uri, string bytes, bool db)")
  fmt.Fprintln(os.Stderr, "  bool tPush(string fromId, string tid, bool readfeed, string uri, string bytes, bool db, bool queue)")
  fmt.Fprintln(os.Stderr, "  bool tDirty(string tid)")
  fmt.Fprintln(os.Stderr)
  os.Exit(0)
}

type httpHeaders map[string]string

func (h httpHeaders) String() string {
  var m map[string]string = h
  return fmt.Sprintf("%s", m)
}

func (h httpHeaders) Set(value string) error {
  parts := strings.Split(value, ": ")
  if len(parts) != 2 {
    return fmt.Errorf("header should be of format 'Key: Value'")
  }
  h[parts[0]] = parts[1]
  return nil
}

func main() {
  flag.Usage = Usage
  var host string
  var port int
  var protocol string
  var urlString string
  var framed bool
  var useHttp bool
  headers := make(httpHeaders)
  var parsedUrl *url.URL
  var trans thrift.TTransport
  _ = strconv.Atoi
  _ = math.Abs
  flag.Usage = Usage
  flag.StringVar(&host, "h", "localhost", "Specify host and port")
  flag.IntVar(&port, "p", 9090, "Specify port")
  flag.StringVar(&protocol, "P", "binary", "Specify the protocol (binary, compact, simplejson, json)")
  flag.StringVar(&urlString, "u", "", "Specify the url")
  flag.BoolVar(&framed, "framed", false, "Use framed transport")
  flag.BoolVar(&useHttp, "http", false, "Use http")
  flag.Var(headers, "H", "Headers to set on the http(s) request (e.g. -H \"Key: Value\")")
  flag.Parse()
  
  if len(urlString) > 0 {
    var err error
    parsedUrl, err = url.Parse(urlString)
    if err != nil {
      fmt.Fprintln(os.Stderr, "Error parsing URL: ", err)
      flag.Usage()
    }
    host = parsedUrl.Host
    useHttp = len(parsedUrl.Scheme) <= 0 || parsedUrl.Scheme == "http" || parsedUrl.Scheme == "https"
  } else if useHttp {
    _, err := url.Parse(fmt.Sprint("http://", host, ":", port))
    if err != nil {
      fmt.Fprintln(os.Stderr, "Error parsing URL: ", err)
      flag.Usage()
    }
  }
  
  cmd := flag.Arg(0)
  var err error
  var cfg *thrift.TConfiguration = nil
  if useHttp {
    trans, err = thrift.NewTHttpClient(parsedUrl.String())
    if len(headers) > 0 {
      httptrans := trans.(*thrift.THttpClient)
      for key, value := range headers {
        httptrans.SetHeader(key, value)
      }
    }
  } else {
    portStr := fmt.Sprint(port)
    if strings.Contains(host, ":") {
           host, portStr, err = net.SplitHostPort(host)
           if err != nil {
                   fmt.Fprintln(os.Stderr, "error with host:", err)
                   os.Exit(1)
           }
    }
    trans = thrift.NewTSocketConf(net.JoinHostPort(host, portStr), cfg)
    if err != nil {
      fmt.Fprintln(os.Stderr, "error resolving address:", err)
      os.Exit(1)
    }
    if framed {
      trans = thrift.NewTFramedTransportConf(trans, cfg)
    }
  }
  if err != nil {
    fmt.Fprintln(os.Stderr, "Error creating transport", err)
    os.Exit(1)
  }
  defer trans.Close()
  var protocolFactory thrift.TProtocolFactory
  switch protocol {
  case "compact":
    protocolFactory = thrift.NewTCompactProtocolFactoryConf(cfg)
    break
  case "simplejson":
    protocolFactory = thrift.NewTSimpleJSONProtocolFactoryConf(cfg)
    break
  case "json":
    protocolFactory = thrift.NewTJSONProtocolFactory()
    break
  case "binary", "":
    protocolFactory = thrift.NewTBinaryProtocolFactoryConf(cfg)
    break
  default:
    fmt.Fprintln(os.Stderr, "Invalid protocol specified: ", protocol)
    Usage()
    os.Exit(1)
  }
  iprot := protocolFactory.GetProtocol(trans)
  oprot := protocolFactory.GetProtocol(trans)
  client := gw.NewGatewayClient(thrift.NewTStandardClient(iprot, oprot))
  if err := trans.Open(); err != nil {
    fmt.Fprintln(os.Stderr, "Error opening socket to ", host, ":", port, " ", err)
    os.Exit(1)
  }
  
  switch cmd {
  case "close":
    if flag.NArg() - 1 != 2 {
      fmt.Fprintln(os.Stderr, "Close requires 2 args")
      flag.Usage()
    }
    argvalue0, err80 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err80 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1 := flag.Arg(2)
    value1 := argvalue1
    fmt.Print(client.Close(context.Background(), value0, value1))
    fmt.Print("\n")
    break
  case "kick":
    if flag.NArg() - 1 != 2 {
      fmt.Fprintln(os.Stderr, "Kick requires 2 args")
      flag.Usage()
    }
    argvalue0, err82 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err82 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1 := []byte(flag.Arg(2))
    value1 := argvalue1
    fmt.Print(client.Kick(context.Background(), value0, value1))
    fmt.Print("\n")
    break
  case "rid":
    if flag.NArg() - 1 != 3 {
      fmt.Fprintln(os.Stderr, "Rid requires 3 args")
      flag.Usage()
    }
    argvalue0, err84 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err84 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1 := flag.Arg(2)
    value1 := argvalue1
    tmp2, err86 := (strconv.Atoi(flag.Arg(3)))
    if err86 != nil {
      Usage()
      return
    }
    argvalue2 := int32(tmp2)
    value2 := argvalue2
    fmt.Print(client.Rid(context.Background(), value0, value1, value2))
    fmt.Print("\n")
    break
  case "rids":
    if flag.NArg() - 1 != 2 {
      fmt.Fprintln(os.Stderr, "Rids requires 2 args")
      flag.Usage()
    }
    argvalue0, err87 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err87 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    arg88 := flag.Arg(2)
    mbTrans89 := thrift.NewTMemoryBufferLen(len(arg88))
    defer mbTrans89.Close()
    _, err90 := mbTrans89.WriteString(arg88)
    if err90 != nil { 
      Usage()
      return
    }
    factory91 := thrift.NewTJSONProtocolFactory()
    jsProt92 := factory91.GetProtocol(mbTrans89)
    containerStruct1 := gw.NewGatewayRidsArgs()
    err93 := containerStruct1.ReadField2(context.Background(), jsProt92)
    if err93 != nil {
      Usage()
      return
    }
    argvalue1 := containerStruct1.Rids
    value1 := argvalue1
    fmt.Print(client.Rids(context.Background(), value0, value1))
    fmt.Print("\n")
    break
  case "push":
    if flag.NArg() - 1 != 3 {
      fmt.Fprintln(os.Stderr, "Push requires 3 args")
      flag.Usage()
    }
    argvalue0, err94 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err94 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1 := flag.Arg(2)
    value1 := argvalue1
    argvalue2 := []byte(flag.Arg(3))
    value2 := argvalue2
    fmt.Print(client.Push(context.Background(), value0, value1, value2))
    fmt.Print("\n")
    break
  case "gPush":
    if flag.NArg() - 1 != 6 {
      fmt.Fprintln(os.Stderr, "GPush requires 6 args")
      flag.Usage()
    }
    argvalue0 := flag.Arg(1)
    value0 := argvalue0
    argvalue1 := flag.Arg(2)
    value1 := argvalue1
    argvalue2 := []byte(flag.Arg(3))
    value2 := argvalue2
    tmp3, err100 := (strconv.Atoi(flag.Arg(4)))
    if err100 != nil {
      Usage()
      return
    }
    argvalue3 := int32(tmp3)
    value3 := argvalue3
    argvalue4 := flag.Arg(5)
    value4 := argvalue4
    argvalue5 := flag.Arg(6) == "true"
    value5 := argvalue5
    fmt.Print(client.GPush(context.Background(), value0, value1, value2, value3, value4, value5))
    fmt.Print("\n")
    break
  case "gConn":
    if flag.NArg() - 1 != 3 {
      fmt.Fprintln(os.Stderr, "GConn requires 3 args")
      flag.Usage()
    }
    argvalue0, err103 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err103 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1 := flag.Arg(2)
    value1 := argvalue1
    argvalue2 := flag.Arg(3)
    value2 := argvalue2
    fmt.Print(client.GConn(context.Background(), value0, value1, value2))
    fmt.Print("\n")
    break
  case "gDisc":
    if flag.NArg() - 1 != 4 {
      fmt.Fprintln(os.Stderr, "GDisc requires 4 args")
      flag.Usage()
    }
    argvalue0, err106 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err106 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1 := flag.Arg(2)
    value1 := argvalue1
    argvalue2 := flag.Arg(3)
    value2 := argvalue2
    tmp3, err109 := (strconv.Atoi(flag.Arg(4)))
    if err109 != nil {
      Usage()
      return
    }
    argvalue3 := int32(tmp3)
    value3 := argvalue3
    fmt.Print(client.GDisc(context.Background(), value0, value1, value2, value3))
    fmt.Print("\n")
    break
  case "gLasts":
    if flag.NArg() - 1 != 4 {
      fmt.Fprintln(os.Stderr, "GLasts requires 4 args")
      flag.Usage()
    }
    argvalue0 := flag.Arg(1)
    value0 := argvalue0
    argvalue1, err111 := (strconv.ParseInt(flag.Arg(2), 10, 64))
    if err111 != nil {
      Usage()
      return
    }
    value1 := argvalue1
    argvalue2 := flag.Arg(3)
    value2 := argvalue2
    argvalue3, err113 := (strconv.ParseInt(flag.Arg(4), 10, 64))
    if err113 != nil {
      Usage()
      return
    }
    value3 := argvalue3
    fmt.Print(client.GLasts(context.Background(), value0, value1, value2, value3))
    fmt.Print("\n")
    break
  case "send":
    if flag.NArg() - 1 != 5 {
      fmt.Fprintln(os.Stderr, "Send requires 5 args")
      flag.Usage()
    }
    argvalue0 := flag.Arg(1)
    value0 := argvalue0
    argvalue1 := flag.Arg(2)
    value1 := argvalue1
    argvalue2 := flag.Arg(3)
    value2 := argvalue2
    argvalue3 := []byte(flag.Arg(4))
    value3 := argvalue3
    argvalue4 := flag.Arg(5) == "true"
    value4 := argvalue4
    fmt.Print(client.Send(context.Background(), value0, value1, value2, value3, value4))
    fmt.Print("\n")
    break
  case "tPush":
    if flag.NArg() - 1 != 7 {
      fmt.Fprintln(os.Stderr, "TPush requires 7 args")
      flag.Usage()
    }
    argvalue0 := flag.Arg(1)
    value0 := argvalue0
    argvalue1 := flag.Arg(2)
    value1 := argvalue1
    argvalue2 := flag.Arg(3) == "true"
    value2 := argvalue2
    argvalue3 := flag.Arg(4)
    value3 := argvalue3
    argvalue4 := []byte(flag.Arg(5))
    value4 := argvalue4
    argvalue5 := flag.Arg(6) == "true"
    value5 := argvalue5
    argvalue6 := flag.Arg(7) == "true"
    value6 := argvalue6
    fmt.Print(client.TPush(context.Background(), value0, value1, value2, value3, value4, value5, value6))
    fmt.Print("\n")
    break
  case "tDirty":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "TDirty requires 1 args")
      flag.Usage()
    }
    argvalue0 := flag.Arg(1)
    value0 := argvalue0
    fmt.Print(client.TDirty(context.Background(), value0))
    fmt.Print("\n")
    break
  case "":
    Usage()
    break
  default:
    fmt.Fprintln(os.Stderr, "Invalid function ", cmd)
  }
}
