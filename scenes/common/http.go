package common

import (
	"context"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"net/http"
	"time"
)

type HttpAppContainer[T scene.Application] struct {
	manager scene.ApplicationManager[T]
	factory scene.ApplicationFactory[T]
	logger  logger.ILogger
	addr    string
	handler http.Handler
	server  *http.Server
	status  scene.AppContainerStatus
}

func NewHttpAppContainer[T scene.Application](
	manager scene.ApplicationManager[T],
	factory scene.ApplicationFactory[T],
	logger logger.ILogger,
	addr string,
	handler http.Handler) *HttpAppContainer[T] {
	hac := &HttpAppContainer[T]{
		manager: manager,
		factory: factory,
		status:  scene.AppContainerStatusStopped,
		addr:    addr,
		handler: handler,
		logger:  logger,
	}
	return hac
}

func (h *HttpAppContainer[T]) Name() string {
	return "scene.app-container.http"
}

func (h *HttpAppContainer[T]) GetAppInfo(appID string) scene.Application {
	return h.manager.GetApp(appID)
}

func (h *HttpAppContainer[T]) ListAppNames() []string {
	return h.manager.ListAppNames()
}

func (h *HttpAppContainer[T]) Status() scene.AppContainerStatus {
	return h.status
}

func (h *HttpAppContainer[T]) Start() error {
	if h.status == scene.AppContainerStatusRunning {
		return nil
	}
	if err := h.startApps(); err != nil {
		h.status = scene.AppContainerStatusError
		return err
	}
	h.server = &http.Server{
		Addr:    h.addr,
		Handler: h.handler,
	}
	h.status = scene.AppContainerStatusRunning
	go func() {
		h.logger.Infof("http server started, listen on %s", h.addr)
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.logger.Errorf("listen: %s\n", err)
			h.status = scene.AppContainerStatusError
		} else {
			h.status = scene.AppContainerStatusStopped
		}
	}()
	return nil
}

func (h *HttpAppContainer[T]) Stop(ctx context.Context) error {
	if h.status != scene.AppContainerStatusRunning {
		return nil
	}

	if err := h.stopApps(); err != nil {
		return err
	}

	subctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := h.server.Shutdown(subctx); err != nil {
		h.logger.Infof("Server Shutdown:", err)
		return err
	}
	return nil
}

func (c *HttpAppContainer[T]) startApps() error {
	for _, app := range c.manager.ListApps() {
		if err := c.factory.Create(app); err != nil {
			c.logger.Errorf("app %s failed to create: %s", app.Name(), err.Error())
		} else {
			c.logger.Infof("app %s created", app.Name())
		}
	}
	return nil
}

func (c *HttpAppContainer[T]) stopApps() error {
	for _, app := range c.manager.ListApps() {
		if err := c.factory.Destroy(app); err != nil {
			c.logger.Errorf("app %s failed to destroy: %s", app.Name(), err.Error())
		} else {
			c.logger.Infof("app %s destroyed", app.Name())
		}
	}
	return nil
}
