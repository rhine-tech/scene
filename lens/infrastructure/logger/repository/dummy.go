package repository

import (
	"fmt"
	"github.com/aynakeya/scene/lens/infrastructure/logger"
)

type DummyLogger struct {
}

func (l *DummyLogger) Debug(args ...interface{}) {
	fmt.Println(args...)
}

func (l *DummyLogger) Debugf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (l *DummyLogger) Info(args ...interface{}) {
	fmt.Println(args...)
}

func (l *DummyLogger) Infof(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (l *DummyLogger) Warn(args ...interface{}) {
	fmt.Println(args...)
}

func (l *DummyLogger) Warnf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (l *DummyLogger) Error(args ...interface{}) {
	fmt.Println(args...)
}

func (l *DummyLogger) Errorf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (l *DummyLogger) WithPrefix(prefix string) logger.ILogger {
	return l
}

func (l *DummyLogger) SetLogLevel(level logger.LogLevel) {

}
