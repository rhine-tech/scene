package gin

import (
	"fmt"
	"time"

	ginCors "github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
)

// GinOption configures the Gin engine before applications are mounted.
type GinOption func(engine *gin.Engine) error

func _ginLogger(log logger.ILogger) gin.HandlerFunc {
	log = registry.Use(log).WithPrefix(scene.NewSceneImplNameNoVer("gin", "router").Identifier())

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		c.Next()

		latency := time.Now().Sub(start)

		msg := fmt.Sprintf("%s %d %s %s \"%s\" (%dms)",
			c.ClientIP(), c.Writer.Status(), c.Request.Method, c.Request.URL.String(),
			c.Errors.ByType(gin.ErrorTypePrivate).Errors(),
			latency.Milliseconds())

		log.Debug(msg)
	}
}

// WithLogger installs Scene's request logger.
func WithLogger(log logger.ILogger) GinOption {
	return func(engine *gin.Engine) error {
		engine.Use(_ginLogger(log))
		return nil
	}
}

// WithCors allows all origins and request headers.
func WithCors() GinOption {
	return func(engine *gin.Engine) error {
		corsCfg := ginCors.DefaultConfig()
		corsCfg.AllowHeaders = []string{"*"}
		corsCfg.AllowAllOrigins = true
		engine.Use(ginCors.New(corsCfg))
		return nil
	}
}

// WithRecovery installs Gin's recovery middleware.
func WithRecovery() GinOption {
	return func(engine *gin.Engine) error {
		engine.Use(gin.Recovery())
		return nil
	}
}

// WithGzip add gzip support for gin engine,
// default level is -1
func WithGzip(level int) GinOption {
	return func(engine *gin.Engine) error {
		engine.Use(gzip.Gzip(level))
		return nil
	}
}

//type ginLogMessage struct {
//	logger.LogMessage
//	Method         string `json:"method"`
//	Path           string `json:"path"`
//	Query          string `json:"query"`
//	SourceIP       string `json:"source_ip"`
//	ResponseStatus int    `json:"response_status"`
//	Latency        int64  `json:"latency"`
//}
//
//func newGinLogger(log logger.ILogger) gin.HandlerFunc {
//	log = log.WithPrefix(scene.NewSceneImplNameNoVer("gin").Identifier())
//	ingestor := registry.AcquireSingleton(ingestion.CommonIngestor(nil)).UsePipe("scene.app-container.http.gin")
//	taskDispatcher := registry.AcquireSingleton(asynctask.TaskDispatcher(nil))
//
//	return func(c *gin.Context) {
//		// Start timer
//		start := time.Now()
//		c.Next()
//
//		latency := time.Now().Sub(start)
//
//		msg := fmt.Sprintf("%s %d %s %s \"%s\" (%dms)",
//			c.ClientIP(), c.Writer.Status(), c.Request.Method, c.Request.URL.String(),
//			c.Errors.ByType(gin.ErrorTypePrivate).String(),
//			latency.Milliseconds())
//
//		log.Info(msg)
//		taskDispatcher.Run(func() error {
//			err := ingestor.Ingest(ginLogMessage{
//				LogMessage: logger.LogMessage{
//					Timestamp: start.Unix(),
//					Level:     logger.LogLevelInfo,
//					Message:   msg,
//				},
//				Method:         c.Request.Method,
//				Path:           c.Request.URL.Path,
//				Query:          c.Request.URL.RawQuery,
//				SourceIP:       c.ClientIP(),
//				ResponseStatus: c.Writer.Status(),
//				Latency:        latency.Milliseconds(),
//			})
//			if err != nil {
//				log.Errorf("failed to ingest log message: %v", err)
//			}
//			return nil
//		})
//	}
//}
