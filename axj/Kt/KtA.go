package Kt

import (
	"os"
	"os/signal"
	"syscall"
)

// 三元表达式
func If(a bool, b, c interface{}) interface{} {
	if a {
		return b
	}

	return c
}

// 关闭信号
func Signal() os.Signal {
	c := make(chan os.Signal, 0)
	signal.Notify(c, syscall.SIGTERM)
	return <-c
}
