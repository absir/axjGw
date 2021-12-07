package PProto

import (
	"axj/Kt/Kt"
	"axj/Kt/KtBuffer"
	"axj/Kt/KtUnsafe"
	"strings"
)

var SERV_NAME_ERR = Kt.NewErrReason("SERV_NAME_ERR")

type HostCtx struct {
	i  int
	si int
}

func hostReadServerName(ctx interface{}, buffer *KtBuffer.Buffer, data []byte, pName *string, host string, hostLen int, servName string, nameFun func(name string) string) (bool, error) {
	buffer.Write(data)
	bs := buffer.Bytes()
	bLen := len(bs)
	hCtx := ctx.(*HostCtx)
	si := hCtx.si
	for i := hCtx.i; i < bLen; i = hCtx.i {
		hCtx.i = i + 1
		b := bs[i]
		if b == '\r' || b == '\n' {
			// CRLF
			if i > si {
				line := KtUnsafe.BytesToString(bs[si:i])
				// println(line)
				lLen := len(line)
				if lLen > hostLen && strings.EqualFold(line[:hostLen], host) {
					name := strings.TrimSpace(line[hostLen:])
					if nameFun != nil {
						name = nameFun(name)
					}

					if servName != "" && !strings.HasSuffix(name, servName) {
						return true, SERV_NAME_ERR
					}

					*pName = name
					return true, nil
				}
			}

			si = hCtx.i
			hCtx.si = si
		}
	}

	return false, nil
}
