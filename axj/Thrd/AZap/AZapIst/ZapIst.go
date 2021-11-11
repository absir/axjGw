package AZapIst

import (
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtCfg"
	"axj/Kt/KtCvt"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitCfg(std bool, opts ...zap.Option) {
	var list *Util.ArrayList = nil
	if std && KtCfg.GetType(APro.Cfg, "log.std", KtCvt.Bool, true).(bool) {
		// 标准扩展
		list = Util.NewArrayList()
		list.Add(zap.AddCaller())
	}

	if list != nil && !list.IsEmpty() {
		oLen := 0
		if opts != nil {
			oLen = len(opts)
		}

		size := list.Size()
		nOpts := make([]zap.Option, list.Size()+oLen)
		for i := list.Size() - 1; i >= 0; i-- {
			nOpts[i] = list.Get(i).(zap.Option)
		}

		for i := 0; i < oLen; i++ {
			nOpts[size+i] = opts[i]
		}

		opts = nOpts
	}

	core := fileCore()
	if core == nil {
		var config zap.Config
		if Kt.Env < Kt.Test {
			// 开发，调试环境日志
			config = zap.NewDevelopmentConfig()

		} else {
			// 测试，正式环境日志
			config = zap.NewProductionConfig()
		}

		APro.SubCfgBind("log", &config)
		AZap.SetLogger(config.Build(opts...))

	} else {
		AZap.SetLogger(zap.New(*core, opts...), nil)
	}
}

func fileCore() *zapcore.Core {
	if !APro.GetCfg("log.file", KtCvt.Bool, Kt.Env >= Kt.Test).(bool) {
		return nil
	}

	// logPath 日志文件路径
	// logLevel 日志级别 debug/info/warn/error
	// maxSize 单个文件大小,MB
	// maxBackups 保存的文件个数
	// maxAge 保存的天数， 没有的话不删除
	// compress 压缩
	// jsonFormat 是否输出为json格式
	// shoowLine 显示代码行
	// logInConsole 是否同时输出到控制台
	encoderConfig := &zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "line",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.ShortCallerEncoder,     // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}
	APro.SubCfgBind("log.encoder", encoderConfig)
	encoder := zapcore.NewConsoleEncoder(*encoderConfig)

	warnLogger := &lumberjack.Logger{
		Filename:   APro.Path() + "/bin/warn.log", // 日志文件路径
		MaxSize:    30,                            // 单个文件大小,MB
		MaxBackups: 30,                            // 最多保留300个备份
	}
	APro.SubCfgBind("log.warn", warnLogger)

	errorLogger := &lumberjack.Logger{
		Filename:   APro.Path() + "/bin/error.log", // 日志文件路径
		MaxSize:    10,                             // 单个文件大小,MB
		MaxBackups: 60,                             // 最多保留300个备份
	}
	APro.SubCfgBind("error.warn", errorLogger)

	// 实现两个判断日志等级的interface
	warnLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl <= zapcore.WarnLevel
	})

	errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl > zapcore.WarnLevel
	})

	//if(APro.GetCfg("log.out", KtCvt.Bool, Kt.Env < Kt.Test).(bool)) {
	//	core := zapcore.NewTee(
	//		zapcore.NewCore(encoder, zapcore.AddSync(warnLogger), warnLevel),
	//		zapcore.NewCore(encoder, zapcore.AddSync(errorLogger), errorLevel),
	//		zapcore.NewCore(encoder, zapcore.AddSync(fmt))
	//	)
	//
	//} else {
	//
	//}

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(warnLogger), warnLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(errorLogger), errorLevel),
	)

	return &core
}
