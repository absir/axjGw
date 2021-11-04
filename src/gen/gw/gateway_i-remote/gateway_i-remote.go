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
  fmt.Fprintln(os.Stderr, "  Result close(i64 cid, string reason)")
  fmt.Fprintln(os.Stderr, "  Result kick(i64 cid, string bytes)")
  fmt.Fprintln(os.Stderr, "  Result conn(i64 cid, string gid, string unique)")
  fmt.Fprintln(os.Stderr, "  void disc(i64 cid, string gid, string unique, i32 connVer)")
  fmt.Fprintln(os.Stderr, "  Result alive(i64 cid)")
  fmt.Fprintln(os.Stderr, "  Result rid(i64 cid, string name, i32 rid)")
  fmt.Fprintln(os.Stderr, "  Result rids(i64 cid,  rids)")
  fmt.Fprintln(os.Stderr, "  Result last(i64 cid, string gid, i32 connVer, bool continuous)")
  fmt.Fprintln(os.Stderr, "  Result push(i64 cid, string uri, string bytes, bool isolate, i64 id)")
  fmt.Fprintln(os.Stderr, "  Result gQueue(string gid, i64 cid, string unique, bool clear)")
  fmt.Fprintln(os.Stderr, "  Result gClear(string gid, bool queue, bool last)")
  fmt.Fprintln(os.Stderr, "  Result gLasts(string gid, i64 cid, string unique, i64 lastId, bool continuous)")
  fmt.Fprintln(os.Stderr, "  i64 gPush(string gid, string uri, string bytes, bool isolate, i32 qs, bool queue, string unique, i64 fid)")
  fmt.Fprintln(os.Stderr, "  Result gPushA(string gid, i64 id, bool succ)")
  fmt.Fprintln(os.Stderr, "  Result send(string fromId, string toId, string uri, string bytes, bool db)")
  fmt.Fprintln(os.Stderr, "  Result tPush(string fromId, string tid, bool readfeed, string uri, string bytes, bool db, bool queue)")
  fmt.Fprintln(os.Stderr, "  Result tDirty(string tid)")
  fmt.Fprintln(os.Stderr, "  Result tStarts(string tid)")
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
  client := gw.NewGatewayIClient(thrift.NewTStandardClient(iprot, oprot))
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
    argvalue0, err56 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err56 != nil {
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
    argvalue0, err58 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err58 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1 := []byte(flag.Arg(2))
    value1 := argvalue1
    fmt.Print(client.Kick(context.Background(), value0, value1))
    fmt.Print("\n")
    break
  case "conn":
    if flag.NArg() - 1 != 3 {
      fmt.Fprintln(os.Stderr, "Conn requires 3 args")
      flag.Usage()
    }
    argvalue0, err60 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err60 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1 := flag.Arg(2)
    value1 := argvalue1
    argvalue2 := flag.Arg(3)
    value2 := argvalue2
    fmt.Print(client.Conn(context.Background(), value0, value1, value2))
    fmt.Print("\n")
    break
  case "disc":
    if flag.NArg() - 1 != 4 {
      fmt.Fprintln(os.Stderr, "Disc requires 4 args")
      flag.Usage()
    }
    argvalue0, err63 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err63 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1 := flag.Arg(2)
    value1 := argvalue1
    argvalue2 := flag.Arg(3)
    value2 := argvalue2
    tmp3, err66 := (strconv.Atoi(flag.Arg(4)))
    if err66 != nil {
      Usage()
      return
    }
    argvalue3 := int32(tmp3)
    value3 := argvalue3
    fmt.Print(client.Disc(context.Background(), value0, value1, value2, value3))
    fmt.Print("\n")
    break
  case "alive":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "Alive requires 1 args")
      flag.Usage()
    }
    argvalue0, err67 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err67 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    fmt.Print(client.Alive(context.Background(), value0))
    fmt.Print("\n")
    break
  case "rid":
    if flag.NArg() - 1 != 3 {
      fmt.Fprintln(os.Stderr, "Rid requires 3 args")
      flag.Usage()
    }
    argvalue0, err68 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err68 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1 := flag.Arg(2)
    value1 := argvalue1
    tmp2, err70 := (strconv.Atoi(flag.Arg(3)))
    if err70 != nil {
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
    argvalue0, err71 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err71 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    arg72 := flag.Arg(2)
    mbTrans73 := thrift.NewTMemoryBufferLen(len(arg72))
    defer mbTrans73.Close()
    _, err74 := mbTrans73.WriteString(arg72)
    if err74 != nil { 
      Usage()
      return
    }
    factory75 := thrift.NewTJSONProtocolFactory()
    jsProt76 := factory75.GetProtocol(mbTrans73)
    containerStruct1 := gw.NewGatewayIRidsArgs()
    err77 := containerStruct1.ReadField2(context.Background(), jsProt76)
    if err77 != nil {
      Usage()
      return
    }
    argvalue1 := containerStruct1.Rids
    value1 := argvalue1
    fmt.Print(client.Rids(context.Background(), value0, value1))
    fmt.Print("\n")
    break
  case "last":
    if flag.NArg() - 1 != 4 {
      fmt.Fprintln(os.Stderr, "Last requires 4 args")
      flag.Usage()
    }
    argvalue0, err78 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err78 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1 := flag.Arg(2)
    value1 := argvalue1
    tmp2, err80 := (strconv.Atoi(flag.Arg(3)))
    if err80 != nil {
      Usage()
      return
    }
    argvalue2 := int32(tmp2)
    value2 := argvalue2
    argvalue3 := flag.Arg(4) == "true"
    value3 := argvalue3
    fmt.Print(client.Last(context.Background(), value0, value1, value2, value3))
    fmt.Print("\n")
    break
  case "push":
    if flag.NArg() - 1 != 5 {
      fmt.Fprintln(os.Stderr, "Push requires 5 args")
      flag.Usage()
    }
    argvalue0, err82 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err82 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1 := flag.Arg(2)
    value1 := argvalue1
    argvalue2 := []byte(flag.Arg(3))
    value2 := argvalue2
    argvalue3 := flag.Arg(4) == "true"
    value3 := argvalue3
    argvalue4, err86 := (strconv.ParseInt(flag.Arg(5), 10, 64))
    if err86 != nil {
      Usage()
      return
    }
    value4 := argvalue4
    fmt.Print(client.Push(context.Background(), value0, value1, value2, value3, value4))
    fmt.Print("\n")
    break
  case "gQueue":
    if flag.NArg() - 1 != 4 {
      fmt.Fprintln(os.Stderr, "GQueue requires 4 args")
      flag.Usage()
    }
    argvalue0 := flag.Arg(1)
    value0 := argvalue0
    argvalue1, err88 := (strconv.ParseInt(flag.Arg(2), 10, 64))
    if err88 != nil {
      Usage()
      return
    }
    value1 := argvalue1
    argvalue2 := flag.Arg(3)
    value2 := argvalue2
    argvalue3 := flag.Arg(4) == "true"
    value3 := argvalue3
    fmt.Print(client.GQueue(context.Background(), value0, value1, value2, value3))
    fmt.Print("\n")
    break
  case "gClear":
    if flag.NArg() - 1 != 3 {
      fmt.Fprintln(os.Stderr, "GClear requires 3 args")
      flag.Usage()
    }
    argvalue0 := flag.Arg(1)
    value0 := argvalue0
    argvalue1 := flag.Arg(2) == "true"
    value1 := argvalue1
    argvalue2 := flag.Arg(3) == "true"
    value2 := argvalue2
    fmt.Print(client.GClear(context.Background(), value0, value1, value2))
    fmt.Print("\n")
    break
  case "gLasts":
    if flag.NArg() - 1 != 5 {
      fmt.Fprintln(os.Stderr, "GLasts requires 5 args")
      flag.Usage()
    }
    argvalue0 := flag.Arg(1)
    value0 := argvalue0
    argvalue1, err95 := (strconv.ParseInt(flag.Arg(2), 10, 64))
    if err95 != nil {
      Usage()
      return
    }
    value1 := argvalue1
    argvalue2 := flag.Arg(3)
    value2 := argvalue2
    argvalue3, err97 := (strconv.ParseInt(flag.Arg(4), 10, 64))
    if err97 != nil {
      Usage()
      return
    }
    value3 := argvalue3
    argvalue4 := flag.Arg(5) == "true"
    value4 := argvalue4
    fmt.Print(client.GLasts(context.Background(), value0, value1, value2, value3, value4))
    fmt.Print("\n")
    break
  case "gPush":
    if flag.NArg() - 1 != 8 {
      fmt.Fprintln(os.Stderr, "GPush requires 8 args")
      flag.Usage()
    }
    argvalue0 := flag.Arg(1)
    value0 := argvalue0
    argvalue1 := flag.Arg(2)
    value1 := argvalue1
    argvalue2 := []byte(flag.Arg(3))
    value2 := argvalue2
    argvalue3 := flag.Arg(4) == "true"
    value3 := argvalue3
    tmp4, err103 := (strconv.Atoi(flag.Arg(5)))
    if err103 != nil {
      Usage()
      return
    }
    argvalue4 := int32(tmp4)
    value4 := argvalue4
    argvalue5 := flag.Arg(6) == "true"
    value5 := argvalue5
    argvalue6 := flag.Arg(7)
    value6 := argvalue6
    argvalue7, err106 := (strconv.ParseInt(flag.Arg(8), 10, 64))
    if err106 != nil {
      Usage()
      return
    }
    value7 := argvalue7
    fmt.Print(client.GPush(context.Background(), value0, value1, value2, value3, value4, value5, value6, value7))
    fmt.Print("\n")
    break
  case "gPushA":
    if flag.NArg() - 1 != 3 {
      fmt.Fprintln(os.Stderr, "GPushA requires 3 args")
      flag.Usage()
    }
    argvalue0 := flag.Arg(1)
    value0 := argvalue0
    argvalue1, err108 := (strconv.ParseInt(flag.Arg(2), 10, 64))
    if err108 != nil {
      Usage()
      return
    }
    value1 := argvalue1
    argvalue2 := flag.Arg(3) == "true"
    value2 := argvalue2
    fmt.Print(client.GPushA(context.Background(), value0, value1, value2))
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
  case "tStarts":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "TStarts requires 1 args")
      flag.Usage()
    }
    argvalue0 := flag.Arg(1)
    value0 := argvalue0
    fmt.Print(client.TStarts(context.Background(), value0))
    fmt.Print("\n")
    break
  case "":
    Usage()
    break
  default:
    fmt.Fprintln(os.Stderr, "Invalid function ", cmd)
  }
}
