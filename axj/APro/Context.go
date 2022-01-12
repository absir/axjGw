package APro

import (
	"axj/Kt/Kt"
	"axj/Kt/KtCvt"
	"net"
)

var localIpP *string

func GetLocalIp() string {
	if localIpP == nil {
		localIp := ""
		if Cfg != nil {
			localIp = KtCvt.ToString(Cfg.Get("LocalIp"))
		}

		if localIp == "" {
			ifaces, err := net.Interfaces()
			Kt.Panic(err)
			for _, iface := range ifaces {
				addrs, _ := iface.Addrs()
				if addrs != nil {
					for _, addr := range addrs {
						if ip, ok := addr.(*net.IPNet); ok {
							if ip.Mask[0] != 0xff || ip.Mask[1] != 0xff {
								continue
							}

							if ip4 := ip.IP.To4(); ip4 != nil {
								if ip4[0] == 127 {
									continue
								}

								localIp = ip4.String()
								break
							}
						}
					}
				}

				if localIp != "" {
					break
				}
			}
		}

		localIpP = &localIp
	}

	return *localIpP
}
