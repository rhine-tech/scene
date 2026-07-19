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
	DebugW(message string, keysAndValues ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	InfoW(message string, keysAndValues ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	WarnW(message string, keysAndValues ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	ErrorW(message string, keysAndValues ...interface{})
	WithPrefix(prefix string) ILogger
	SetLogLevel(level LogLevel)
	WithOptions(opts ...Option) ILogger // new api, can replace WithPrefix and SetLogLevel in the future
}

type LogMessage struct {
	Timestamp int64
	Level     LogLevel
	Prefix    string
	Message   string
	Data      map[string]interface{}
}
