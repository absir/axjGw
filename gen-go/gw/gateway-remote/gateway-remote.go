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
  fmt.Fprintln(os.Stderr, "  bool close(i64 cid, string reason)")
  fmt.Fprintln(os.Stderr, "  bool leave(i64 cid, string bytes)")
  fmt.Fprintln(os.Stderr, "  void rid(i64 cid, string name, i32 rid)")
  fmt.Fprintln(os.Stderr, "  void rids(i64 cid,  rids)")
  fmt.Fprintln(os.Stderr, "  bool push(i64 cid, i64 uid, string sid, string uri, string bytes, i32 qs, string unique)")
  fmt.Fprintln(os.Stderr, "  void pushO(i64 cid, i64 uid, string sid, string uri, string bytes)")
  fmt.Fprintln(os.Stderr, "  void dirty(string sid)")
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
    argvalue0, err55 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err55 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1 := flag.Arg(2)
    value1 := argvalue1
    fmt.Print(client.Close(context.Background(), value0, value1))
    fmt.Print("\n")
    break
  case "leave":
    if flag.NArg() - 1 != 2 {
      fmt.Fprintln(os.Stderr, "Leave requires 2 args")
      flag.Usage()
    }
    argvalue0, err57 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err57 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1 := []byte(flag.Arg(2))
    value1 := argvalue1
    fmt.Print(client.Leave(context.Background(), value0, value1))
    fmt.Print("\n")
    break
  case "rid":
    if flag.NArg() - 1 != 3 {
      fmt.Fprintln(os.Stderr, "Rid requires 3 args")
      flag.Usage()
    }
    argvalue0, err59 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err59 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1 := flag.Arg(2)
    value1 := argvalue1
    tmp2, err61 := (strconv.Atoi(flag.Arg(3)))
    if err61 != nil {
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
    argvalue0, err62 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err62 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    arg63 := flag.Arg(2)
    mbTrans64 := thrift.NewTMemoryBufferLen(len(arg63))
    defer mbTrans64.Close()
    _, err65 := mbTrans64.WriteString(arg63)
    if err65 != nil { 
      Usage()
      return
    }
    factory66 := thrift.NewTJSONProtocolFactory()
    jsProt67 := factory66.GetProtocol(mbTrans64)
    containerStruct1 := gw.NewGatewayRidsArgs()
    err68 := containerStruct1.ReadField2(context.Background(), jsProt67)
    if err68 != nil {
      Usage()
      return
    }
    argvalue1 := containerStruct1.Rids
    value1 := argvalue1
    fmt.Print(client.Rids(context.Background(), value0, value1))
    fmt.Print("\n")
    break
  case "push":
    if flag.NArg() - 1 != 7 {
      fmt.Fprintln(os.Stderr, "Push requires 7 args")
      flag.Usage()
    }
    argvalue0, err69 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err69 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1, err70 := (strconv.ParseInt(flag.Arg(2), 10, 64))
    if err70 != nil {
      Usage()
      return
    }
    value1 := argvalue1
    argvalue2 := flag.Arg(3)
    value2 := argvalue2
    argvalue3 := flag.Arg(4)
    value3 := argvalue3
    argvalue4 := []byte(flag.Arg(5))
    value4 := argvalue4
    tmp5, err74 := (strconv.Atoi(flag.Arg(6)))
    if err74 != nil {
      Usage()
      return
    }
    argvalue5 := int32(tmp5)
    value5 := argvalue5
    argvalue6 := flag.Arg(7)
    value6 := argvalue6
    fmt.Print(client.Push(context.Background(), value0, value1, value2, value3, value4, value5, value6))
    fmt.Print("\n")
    break
  case "pushO":
    if flag.NArg() - 1 != 5 {
      fmt.Fprintln(os.Stderr, "PushO requires 5 args")
      flag.Usage()
    }
    argvalue0, err76 := (strconv.ParseInt(flag.Arg(1), 10, 64))
    if err76 != nil {
      Usage()
      return
    }
    value0 := argvalue0
    argvalue1, err77 := (strconv.ParseInt(flag.Arg(2), 10, 64))
    if err77 != nil {
      Usage()
      return
    }
    value1 := argvalue1
    argvalue2 := flag.Arg(3)
    value2 := argvalue2
    argvalue3 := flag.Arg(4)
    value3 := argvalue3
    argvalue4 := []byte(flag.Arg(5))
    value4 := argvalue4
    fmt.Print(client.PushO(context.Background(), value0, value1, value2, value3, value4))
    fmt.Print("\n")
    break
  case "dirty":
    if flag.NArg() - 1 != 1 {
      fmt.Fprintln(os.Stderr, "Dirty requires 1 args")
      flag.Usage()
    }
    argvalue0 := flag.Arg(1)
    value0 := argvalue0
    fmt.Print(client.Dirty(context.Background(), value0))
    fmt.Print("\n")
    break
  case "":
    Usage()
    break
  default:
    fmt.Fprintln(os.Stderr, "Invalid function ", cmd)
  }
}
