package repository

import (
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
	"go.uber.org/zap"
)

// TODO: Implement Zap Logger Core

type zapLoggerImpl struct {
	*zap.SugaredLogger
}

func NewZapLogger() logger.ILogger {
	zapLog, _ := zap.NewProduction()
	sugar := zapLog.Sugar()
	return &zapLoggerImpl{SugaredLogger: sugar}
}

func (z *zapLoggerImpl) Dispose() error {
	_ = z.Sync()
	return nil
}

func (z *zapLoggerImpl) WithPrefix(prefix string) logger.ILogger {
	return z
}

func (z *zapLoggerImpl) SetLogLevel(level logger.LogLevel) {
	return
}
