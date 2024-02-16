package logger

type LogLevel uint32

const (
	LogLevelError LogLevel = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

type LogField map[string]interface{}

func (f LogField) Flatten() []interface{} {
	var res []interface{}
	for k, v := range f {
		res = append(res, k, v)
	}
	return res
}

type ILogger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	DebugW(message string, field LogField)
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	InfoW(message string, field LogField)
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	WarnW(message string, field LogField)
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	ErrorW(message string, field LogField)
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
