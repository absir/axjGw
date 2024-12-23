package nps

type npsConfig struct {
	AdminPort     int    // 管理端口号
	AdminName     string // 管理账号
	AdminPassword string // 管理密码
	HttpPort      int    // http端口
	RtspPort      int    // rtsp端口
}

var NpsConfig = &npsConfig{
	AdminPort:     8782,
	AdminName:     "admin",
	AdminPassword: "",
	HttpPort:      82,
	RtspPort:      83,
}
