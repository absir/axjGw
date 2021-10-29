// Code generated by Thrift Compiler (0.14.2). DO NOT EDIT.

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
	"github.com/apache/thrift/lib/go/thrift"
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
  fmt.Fprintln(os.Stderr, "  Result last(i64 cid, string gid, i32 connVer)")
  fmt.Fprintln(os.Stderr, "  Result push(i64 cid, string uri, string bytes, bool isolate, i64 id)")
  fmt.Fprintln(os.Stderr, "  Result gQueue(string gid, i64 cid, string unique, bool clear)")
  fmt.Fprintln(os.Stderr, "  Result gClear(string gid, bool queue, bool last)")
  fmt.Fprintln(os.Stderr, "  Result gLasts(string gid, i64 cid, string unique, i64 lastId)")
  fmt.Fprintln(os.Stderr, "  i64 gPush(string gid, string uri, string bytes, bool isolate, i32 qs, bool queue, string unique, i64 fid)")
  fmt.Fprintln(os.Stderr, "  Result gPushA(string gid, i64 id, bool succ)")
  fmt.Fprintln(os.Stderr, "  Result teamDirty(string tid)")
  fmt.Fprintln(os.Stderr, "  Result teamStarts(string tid)")
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
    trans, err = thrift.NewTSocket(net.JoinHostPort(host, portStr))
    if err != nil {
      fmt.Fprintln(os.Stderr, "error resolving address:", err)
      os.Exit(1)
    }
    if framed {
      trans = thrift.NewTFramedTransport(trans)
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
    protocolFactory = thrift.NewTCompactProtocolFactory()
    break
  case "simplejson":
    protocolFactory = thrift.NewTSimpleJSONProtocolFactory()
    break
  case "json":
    protocolFactory = thrift.NewTJSONProtocolFactory()
    break
  case "binary", "":
    protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
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
    argvalue0, err50 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err50 != nil {
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
    argvalue0, err52 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err52 != nil {
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
    argvalue0, err54 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err54 != nil {
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
    argvalue0, err57 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err57 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1 := flag.Arg(2)
    value1 := argvalue1
    argvalue2 := flag.Arg(3)
    value2 := argvalue2
    tmp3, err60 := (strconv.Atoi(flag.Arg(4)))
    if err60 != nil {
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
    argvalue0, err61 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err61 != nil {
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
    argvalue0, err62 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err62 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1 := flag.Arg(2)
    value1 := argvalue1
    tmp2, err64 := (strconv.Atoi(flag.Arg(3)))
    if err64 != nil {
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
    argvalue0, err65 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err65 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    arg66 := flag.Arg(2)
    mbTrans67 := thrift.NewTMemoryBufferLen(len(arg66))
    defer mbTrans67.Close()
    _, err68 := mbTrans67.WriteString(arg66)
    if err68 != nil { 
      Usage()
      return
    }
    factory69 := thrift.NewTJSONProtocolFactory()
    jsProt70 := factory69.GetProtocol(mbTrans67)
    containerStruct1 := gw.NewGatewayIRidsArgs()
    err71 := containerStruct1.ReadField2(context.Background(), jsProt70)
    if err71 != nil {
      Usage()
      return
    }
    argvalue1 := containerStruct1.Rids
    value1 := argvalue1
    fmt.Print(client.Rids(context.Background(), value0, value1))
    fmt.Print("\n")
    break
  case "last":
    if flag.NArg() - 1 != 3 {
      fmt.Fprintln(os.Stderr, "Last requires 3 args")
      flag.Usage()
    }
    argvalue0, err72 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err72 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1 := flag.Arg(2)
    value1 := argvalue1
    tmp2, err74 := (strconv.Atoi(flag.Arg(3)))
    if err74 != nil {
      Usage()
      return
    }
    argvalue2 := int32(tmp2)
    value2 := argvalue2
    fmt.Print(client.Last(context.Background(), value0, value1, value2))
    fmt.Print("\n")
    break
  case "push":
    if flag.NArg() - 1 != 5 {
      fmt.Fprintln(os.Stderr, "Push requires 5 args")
      flag.Usage()
    }
    argvalue0, err75 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err75 != nil {
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
    argvalue4, err79 := (strconv.ParseInt(flag.Arg(5), 10, 64))
    if err79 != nil {
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
    argvalue1, err81 := (strconv.ParseInt(flag.Arg(2), 10, 64))
    if err81 != nil {
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
    if flag.NArg() - 1 != 4 {
      fmt.Fprintln(os.Stderr, "GLasts requires 4 args")
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
    argvalue3, err90 := (strconv.ParseInt(flag.Arg(4), 10, 64))
    if err90 != nil {
      Usage()
      return
    }
    value3 := argvalue3
    fmt.Print(client.GLasts(context.Background(), value0, value1, value2, value3))
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
    tmp4, err95 := (strconv.Atoi(flag.Arg(5)))
    if err95 != nil {
      Usage()
      return
    }
    argvalue4 := int32(tmp4)
    value4 := argvalue4
    argvalue5 := flag.Arg(6) == "true"
    value5 := argvalue5
    argvalue6 := flag.Arg(7)
    value6 := argvalue6
    argvalue7, err98 := (strconv.ParseInt(flag.Arg(8), 10, 64))
    if err98 != nil {
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
    argvalue1, err100 := (strconv.ParseInt(flag.Arg(2), 10, 64))
    if err100 != nil {
      Usage()
      return
    }
    value1 := argvalue1
    argvalue2 := flag.Arg(3) == "true"
    value2 := argvalue2
    fmt.Print(client.GPushA(context.Background(), value0, value1, value2))
    fmt.Print("\n")
    break
  case "teamDirty":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "TeamDirty requires 1 args")
      flag.Usage()
    }
    argvalue0 := flag.Arg(1)
    value0 := argvalue0
    fmt.Print(client.TeamDirty(context.Background(), value0))
    fmt.Print("\n")
    break
  case "teamStarts":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "TeamStarts requires 1 args")
      flag.Usage()
    }
    argvalue0 := flag.Arg(1)
    value0 := argvalue0
    fmt.Print(client.TeamStarts(context.Background(), value0))
    fmt.Print("\n")
    break
  case "":
    Usage()
    break
  default:
    fmt.Fprintln(os.Stderr, "Invalid function ", cmd)
  }
}
