package KtError

import (
	"axj/Thrd/AZap"
	"go.uber.org/zap"
)

func ErrRecover(key string) {
	if err := recover(); err != nil {
		if _err, ok := err.(error); ok {
			AZap.LoggerS.Warn("Convert Err", zap.Namespace(key), zap.Error(_err))

		} else {
			AZap.LoggerS.Warn("Convert Err", zap.Namespace(key), zap.Reflect("error", err))
		}
	}
}
