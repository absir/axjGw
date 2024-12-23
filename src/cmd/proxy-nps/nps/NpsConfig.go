package nps

type npsConfig struct {
	AdminAddr     string // 管理端口号
	AdminName     string // 管理账号
	AdminPassword string // 管理密码
	HttpAddr      string // http端口
	RtspAddr      string // rtsp端口
}

var NpsConfig = &npsConfig{
	AdminAddr:     "0.0.0.0:8782",
	AdminName:     "admin",
	AdminPassword: "",
	HttpAddr:      "0.0.0.0:82",
	RtspAddr:      "0.0.0.0:83",
}
