package AZap

import (
	"axj/Kt/Kt"
	"go.uber.org/zap"
)

var Logger *zap.Logger

func init() {
	logger, err := zap.NewDevelopment(zap.AddCaller())
	Kt.Panic(err)
	Logger = logger
}
