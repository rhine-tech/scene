package repository

import (
	"github.com/mattn/go-colorable"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
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
	{
		enc.AppendString(logger.LogColorMagenta.Add("[" + loggerName + "]"))
	}
}

type zapLoggerImpl struct {
	*zap.SugaredLogger
	level zap.AtomicLevel
}

type namedZapLoggerImpl struct {
	*zap.SugaredLogger
	parent *zapLoggerImpl
}

func (n *namedZapLoggerImpl) WithPrefix(prefix string) logger.ILogger {
	return n.parent.WithPrefix(prefix)
}

func (n *namedZapLoggerImpl) SetLogLevel(level logger.LogLevel) {
	n.parent.SetLogLevel(level)
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
	cfg.ConsoleSeparator = " "
	zapLog := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		zapcore.AddSync(colorable.NewColorableStdout()),
		level,
	))
	sugar := zapLog.Sugar()
	return &zapLoggerImpl{SugaredLogger: sugar, level: level}
}

func (z *zapLoggerImpl) Dispose() error {
	_ = z.Sync()
	return nil
}

func (z *zapLoggerImpl) WithPrefix(prefix string) logger.ILogger {
	return &namedZapLoggerImpl{SugaredLogger: z.SugaredLogger.Named(prefix), parent: z}
}

func (z *zapLoggerImpl) SetLogLevel(level logger.LogLevel) {
	z.level.SetLevel(_logLevelMap[level])
	return
}
