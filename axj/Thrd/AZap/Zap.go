package AZap

import (
	"axj/Kt/Kt"
	"fmt"
	"go.uber.org/zap"
)

var Logger *zap.Logger
var LoggerS *zap.Logger

func init() {
	logger, err := zap.NewDevelopment(zap.AddCaller())
	Kt.Panic(err)
	Logger = logger
	LoggerS = logger.WithOptions(zap.AddCallerSkip(1))
}

func SetLogger(logger *zap.Logger, err error) {
	Kt.Err(err, true)
	if logger != nil {
		Logger = logger
		LoggerS = logger.WithOptions(zap.AddCallerSkip(1))
	}
}

/**
Go 字符串格式化符号:
格  式	描  述
%v	按值的本来值输出
%+v	在 %v 基础上，对结构体字段名和值进行展开
%#v	输出 Go 语言语法格式的值
%T	输出 Go 语言语法格式的类型和值
%%	输出 % 本体
%b	整型以二进制方式显示
%o	整型以八进制方式显示
%d	整型以十进制方式显示
%x	整型以十六进制方式显示
%X	整型以十六进制、字母大写方式显示
%U	Unicode 字符
%f	浮点数
%p	指针，十六进制方式显示
*/
func Debug(msg string, args ...interface{}) {
	if LoggerS.Core().Enabled(zap.DebugLevel) {
		LoggerS.Debug(fmt.Sprintf(msg, args...))
	}
}

func Info(msg string, args ...interface{}) {
	if LoggerS.Core().Enabled(zap.InfoLevel) {
		LoggerS.Info(fmt.Sprintf(msg, args...))
	}
}

func Warn(msg string, args ...interface{}) {
	if LoggerS.Core().Enabled(zap.WarnLevel) {
		LoggerS.Warn(fmt.Sprintf(msg, args...))
	}
}

func Error(msg string, args ...interface{}) {
	if LoggerS.Core().Enabled(zap.ErrorLevel) {
		LoggerS.Error(fmt.Sprintf(msg, args...))
	}
}
