package logger

// NoopLogger discards all log messages. It is intended for dependency
// injection in tests that do not need to inspect logging behavior.
type NoopLogger struct{}

var _ ILogger = NoopLogger{}

func (NoopLogger) Debug(...interface{})            {}
func (NoopLogger) Debugf(string, ...interface{})   {}
func (NoopLogger) DebugW(string, ...interface{})   {}
func (NoopLogger) Info(...interface{})             {}
func (NoopLogger) Infof(string, ...interface{})    {}
func (NoopLogger) InfoW(string, ...interface{})    {}
func (NoopLogger) Warn(...interface{})             {}
func (NoopLogger) Warnf(string, ...interface{})    {}
func (NoopLogger) WarnW(string, ...interface{})    {}
func (NoopLogger) Error(...interface{})            {}
func (NoopLogger) Errorf(string, ...interface{})   {}
func (NoopLogger) ErrorW(string, ...interface{})   {}
func (l NoopLogger) WithPrefix(string) ILogger     { return l }
func (NoopLogger) SetLogLevel(LogLevel)            {}
func (l NoopLogger) WithOptions(...Option) ILogger { return l }
