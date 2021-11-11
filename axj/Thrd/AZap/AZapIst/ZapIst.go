package AZapIst

import (
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtCfg"
	"axj/Kt/KtCvt"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"time"
)

func InitCfg(std bool, opts ...zap.Option) {
	var config zap.Config
	if Kt.Env < Kt.Test {
		// 开发，调试环境日志
		config = zap.NewDevelopmentConfig()

	} else {
		// 测试，正式环境日志
		config = zap.NewProductionConfig()
	}

	APro.SubCfgBind("log", &config)

	var list *Util.ArrayList = nil
	if std && KtCfg.GetType(APro.Cfg, "log.std", KtCvt.Bool, true).(bool) {
		// 标准扩展
		list = Util.NewArrayList()
		list.Add(zap.AddCaller())
		//list.Add(stdCore())
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

	logger, err := config.Build(opts...)
	Kt.Err(err, true)
	if logger != nil {
		AZap.Logger = logger
	}
}

func stdCore() interface{} {
	// 设置一些基本日志格式 具体含义还比较好理解，直接看zap源码也不难懂
	encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey:  "msg",
		LevelKey:    "level",
		EncodeLevel: zapcore.CapitalLevelEncoder,
		TimeKey:     "ts",
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2017-01-01 00:00:00"))
		},
		CallerKey:    "file",
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	})

	// 实现两个判断日志等级的interface
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.WarnLevel
	})

	warnLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.WarnLevel
	})

	// 获取 info、warn日志文件的io.Writer 抽象 getWriter() 在下方实现
	infoWriter := getWriter("/path/log/demo.log")
	warnWriter := getWriter("/path/log/demo_error.log")

	// 最后创建具体的Logger
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(infoWriter), infoLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(warnWriter), warnLevel),
	)

	return core
}

func getWriter(filename string) io.Writer {
	// 生成rotatelogs的Logger 实际生成的文件名 demo.log.YYmmddHH
	// demo.log是指向最新日志的链接
	// 保存7天内的日志，每1小时(整点)分割一次日志
	hook, err := rotatelogs.New(
		filename+".%Y%m%d%H", // 没有使用go风格反人类的format格式
		rotatelogs.WithLinkName(filename),
		rotatelogs.WithMaxAge(time.Hour*24*7),
		rotatelogs.WithRotationTime(time.Hour),
	)

	if err != nil {
		panic(err)
	}

	return hook
}
