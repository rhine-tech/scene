package logger

type LogLevel uint32

const (
	LogLevelError LogLevel = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

type ILogger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	WithPrefix(prefix string) ILogger
	SetLogLevel(level LogLevel)
}

type LogMessage struct {
	Timestamp int64
	Level     LogLevel
	Prefix    string
	Message   string
	Data      map[string]interface{}
}
