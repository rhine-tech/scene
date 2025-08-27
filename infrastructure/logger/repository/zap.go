package repository

import (
	"github.com/mattn/go-colorable"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

var _logLevelMap = map[logger.LogLevel]zapcore.Level{
	logger.LogLevelDebug: zapcore.DebugLevel,
	logger.LogLevelInfo:  zapcore.InfoLevel,
	logger.LogLevelWarn:  zapcore.WarnLevel,
	logger.LogLevelError: zapcore.ErrorLevel,
}

var (
	_levelToCapitalColorString = make(map[zapcore.Level]string, len(logger.LogColorMap))
)

func init() {
	for level, color := range logger.LogColorMap {
		_levelToCapitalColorString[_logLevelMap[level]] = color.Add("[" + _logLevelMap[level].CapitalString() + "]")
	}
}

// reference tp https://blog.sandipb.net/2018/05/03/using-zap-creating-custom-encoders/
func syslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("Jan  2 15:04:05"))
}

func customLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	s, ok := _levelToCapitalColorString[level]
	if !ok {
		s = logger.LogColorRed.Add(level.CapitalString())
	}
	enc.AppendString(s)
}

func customNamedEncoder(loggerName string, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(logger.LogColorMagenta.Add("[" + loggerName + "]"))
}

func customCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(logger.LogColorBlue.Add("[" + caller.TrimmedPath() + "]"))
}

type zapLoggerImpl struct {
	*zap.SugaredLogger
	skip1 *zap.SugaredLogger
	level zap.AtomicLevel
}

type namedZapLoggerImpl struct {
	*zap.SugaredLogger
	skip1  *zap.SugaredLogger
	parent *zapLoggerImpl
}

func (n *namedZapLoggerImpl) WithPrefix(prefix string) logger.ILogger {
	return n.parent.WithPrefix(prefix)
}

func (n *namedZapLoggerImpl) WithOptions(options ...logger.Option) logger.ILogger {
	return n.parent.WithOptions(options...)
}

func (n *namedZapLoggerImpl) SetLogLevel(level logger.LogLevel) {
	n.parent.SetLogLevel(level)
}

func (z *namedZapLoggerImpl) DebugW(message string, keysAndValues ...interface{}) {
	z.skip1.Debugw(message, keysAndValues...)
}

func (z *namedZapLoggerImpl) InfoW(message string, keysAndValues ...interface{}) {
	z.skip1.Infow(message, keysAndValues...)
}

func (z *namedZapLoggerImpl) WarnW(message string, keysAndValues ...interface{}) {
	z.skip1.Warnw(message, keysAndValues...)
}

func (z *namedZapLoggerImpl) ErrorW(message string, keysAndValues ...interface{}) {
	z.skip1.Errorw(message, keysAndValues...)
}

func (z *namedZapLoggerImpl) DebugS(message string, fields logger.LogField) {
	z.skip1.Debugw(message, fields.Flatten()...)
}

func (z *namedZapLoggerImpl) InfoS(message string, fields logger.LogField) {
	z.skip1.Infow(message, fields.Flatten()...)
}

func (z *namedZapLoggerImpl) WarnS(message string, fields logger.LogField) {
	z.skip1.Warnw(message, fields.Flatten()...)
}

func (z *namedZapLoggerImpl) ErrorS(message string, fields logger.LogField) {
	z.skip1.Errorw(message, fields.Flatten()...)
}

func NewZapLogger() logger.ILogger {
	zapLog, _ := zap.NewProduction()
	sugar := zapLog.Sugar()
	return &zapLoggerImpl{SugaredLogger: sugar}
}

func NewZapColoredLogger() logger.ILogger {
	cfg := zap.NewProductionEncoderConfig()
	level := zap.NewAtomicLevel()
	level.SetLevel(zapcore.DebugLevel)
	cfg.EncodeLevel = customLevelEncoder
	cfg.EncodeTime = syslogTimeEncoder
	cfg.EncodeName = customNamedEncoder
	cfg.EncodeCaller = customCallerEncoder
	cfg.ConsoleSeparator = " "
	zapLog := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		zapcore.AddSync(colorable.NewColorableStdout()),
		level,
	), zap.AddCaller(), zap.AddCallerSkip(0))
	sugar := zapLog.Sugar()
	return &zapLoggerImpl{SugaredLogger: sugar, skip1: sugar.WithOptions(zap.AddCallerSkip(1)), level: level}
}

func (z *zapLoggerImpl) Dispose() error {
	_ = z.Sync()
	return nil
}

func (z *zapLoggerImpl) WithPrefix(prefix string) logger.ILogger {
	return &namedZapLoggerImpl{SugaredLogger: z.SugaredLogger.Named(prefix), skip1: z.skip1.Named(prefix), parent: z}
}

func (z *zapLoggerImpl) WithOptions(options ...logger.Option) logger.ILogger {
	named := &namedZapLoggerImpl{SugaredLogger: z.SugaredLogger, skip1: z.skip1, parent: z}
	for _, option := range options {
		switch option.Name {
		case logger.OptionCallerSkip:
			named.SugaredLogger = named.SugaredLogger.WithOptions(zap.AddCallerSkip(option.Value.(int)))
			named.skip1 = z.skip1.WithOptions(zap.AddCallerSkip(option.Value.(int)))
		case logger.OptionWithPrefix:
			named.SugaredLogger = named.SugaredLogger.Named(option.Value.(string))
			named.skip1 = z.skip1.Named(option.Value.(string))
		default:
			// do nothing
		}
	}
	return named
}

func (z *zapLoggerImpl) SetLogLevel(level logger.LogLevel) {
	z.level.SetLevel(_logLevelMap[level])
	return
}

func (z *zapLoggerImpl) DebugW(message string, keysAndValues ...interface{}) {
	z.skip1.Debugw(message, keysAndValues...)
}

func (z *zapLoggerImpl) InfoW(message string, keysAndValues ...interface{}) {
	z.skip1.Infow(message, keysAndValues...)
}

func (z *zapLoggerImpl) WarnW(message string, keysAndValues ...interface{}) {
	z.skip1.Warnw(message, keysAndValues...)
}

func (z *zapLoggerImpl) ErrorW(message string, keysAndValues ...interface{}) {
	z.skip1.Errorw(message, keysAndValues...)
}

func (z *zapLoggerImpl) DebugS(message string, fields logger.LogField) {
	z.skip1.Debugw(message, fields.Flatten()...)
}

func (z *zapLoggerImpl) InfoS(message string, fields logger.LogField) {
	z.skip1.Infow(message, fields.Flatten()...)
}

func (z *zapLoggerImpl) WarnS(message string, fields logger.LogField) {
	z.skip1.Warnw(message, fields.Flatten()...)
}

func (z *zapLoggerImpl) ErrorS(message string, fields logger.LogField) {
	z.skip1.Errorw(message, fields.Flatten()...)
}
