package gin

import (
	"fmt"
	"github.com/aynakeya/scene"
	"github.com/aynakeya/scene/lens/infrastructure/asynctask"
	"github.com/aynakeya/scene/lens/infrastructure/ingestion"
	"github.com/aynakeya/scene/lens/infrastructure/logger"
	"github.com/aynakeya/scene/registry"
	scommon "github.com/aynakeya/scene/scenes/common"
	ginCors "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

func createGinEngine() *gin.Engine {
	if registry.Config.GetBool("scene.debug") {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(gin.Recovery(), newGinLogger(registry.Logger))
	corsCfg := ginCors.DefaultConfig()
	corsCfg.AllowHeaders = []string{"*"}
	corsCfg.AllowAllOrigins = true
	engine.Use(ginCors.New(corsCfg))
	return engine
}

type ginLogMessage struct {
	logger.LogMessage
	Method         string `json:"method"`
	Path           string `json:"path"`
	Query          string `json:"query"`
	SourceIP       string `json:"source_ip"`
	ResponseStatus int    `json:"response_status"`
	Latency        int64  `json:"latency"`
}

func newGinLogger(log logger.ILogger) gin.HandlerFunc {
	log = log.WithPrefix("scene.app-container.http.gin")
	ingestor := registry.AcquireSingleton(ingestion.CommonIngestor(nil)).UsePipe("scene.app-container.http.gin")
	taskDispatcher := registry.AcquireSingleton(asynctask.TaskDispatcher(nil))

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		c.Next()

		latency := time.Now().Sub(start)

		msg := fmt.Sprintf("%s %d %s %s \"%s\" (%dms)",
			c.ClientIP(), c.Writer.Status(), c.Request.Method, c.Request.URL.String(),
			c.Errors.ByType(gin.ErrorTypePrivate).String(),
			latency.Milliseconds())

		log.Info(msg)
		taskDispatcher.Run(func() error {
			err := ingestor.Ingest(ginLogMessage{
				LogMessage: logger.LogMessage{
					Timestamp: start.Unix(),
					Level:     logger.LogLevelInfo,
					Prefix:    SceneName,
					Message:   msg,
				},
				Method:         c.Request.Method,
				Path:           c.Request.URL.Path,
				Query:          c.Request.URL.RawQuery,
				SourceIP:       c.ClientIP(),
				ResponseStatus: c.Writer.Status(),
				Latency:        latency.Milliseconds(),
			})
			if err != nil {
				log.Errorf("failed to ingest log message: %v", err)
			}
			return nil
		})
	}
}

type ginContainer struct {
	*scommon.HttpAppContainer[GinApplication]
}

func (g *ginContainer) Name() string {
	return SceneName
}

func NewAppContainer(addr string, apps ...GinApplication) scene.ApplicationContainer {
	ginEngine := createGinEngine()
	return &ginContainer{scommon.NewHttpAppContainer(
		scommon.NewAppManager(apps...),
		NewAppFactory(ginEngine),
		registry.Logger.WithPrefix(SceneName),
		addr,
		ginEngine,
	)}
}
