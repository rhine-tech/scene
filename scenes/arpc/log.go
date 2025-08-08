package arpc

import (
	"github.com/lesismal/arpc/log"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
)

func init() {
	log.SetLogger(&logDelegate{})
}

type logDelegate struct {
	log logger.ILogger
}

func (l *logDelegate) setupLogger() bool {
	if l.log == nil && registry.Logger != nil {
		l.log = registry.Logger.WithPrefix(scene.NewSceneImplNameNoVer("arpc", "internal").Identifier())
	}
	return l.log != nil
}

func (l *logDelegate) SetLevel(lvl int) {
	if !l.setupLogger() {
		return
	}
	switch lvl {
	case log.LevelAll, log.LevelDebug:
		l.log.SetLogLevel(logger.LogLevelDebug)
	case log.LevelInfo:
		l.log.SetLogLevel(logger.LogLevelInfo)
	case log.LevelWarn:
		l.log.SetLogLevel(logger.LogLevelWarn)
	case log.LevelError:
		l.log.SetLogLevel(logger.LogLevelError)
	default:
		l.log.SetLogLevel(logger.LogLevelError)
	}
}

func (l *logDelegate) Debug(format string, v ...interface{}) {
	if !l.setupLogger() {
		return
	}
	l.log.Debugf(format, v...)
}

func (l *logDelegate) Info(format string, v ...interface{}) {
	if !l.setupLogger() {
		return
	}
	l.log.Infof(format, v...)
}

func (l *logDelegate) Warn(format string, v ...interface{}) {
	if !l.setupLogger() {
		return
	}
	l.log.Warnf(format, v...)
}

func (l *logDelegate) Error(format string, v ...interface{}) {
	if !l.setupLogger() {
		return
	}
	l.log.Errorf(format, v...)
}
