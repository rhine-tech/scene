package engines

import (
	"context"
	"fmt"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	"github.com/rhine-tech/scene/utils"
	"os"
	"os/signal"
	"reflect"
	"strings"
)

var (
	errStartEngineFailed = fmt.Errorf("failed to start engine")
)

type BasicEngine struct {
	logger     logger.ILogger
	containers map[string]scene.ApplicationContainer
}

func NewEngine(logger logger.ILogger, containers ...scene.ApplicationContainer) scene.Engine {
	e := &BasicEngine{
		logger:     logger.WithPrefix("scene.engine"),
		containers: make(map[string]scene.ApplicationContainer),
	}
	for _, container := range containers {
		_ = e.AddContainer(container)
	}
	return e
}

func (eg *BasicEngine) printContainersInfo() {
	containers := eg.ListContainers()

	info := make([]string, len(containers))
	for idx, container := range containers {
		info[idx] = utils.FormatContainerInfo(idx, container)
	}

	eg.logger.Infof("successfully loaded %d containers. \n\n%s",
		len(containers),
		strings.Join(info, "\n"))
}

func (eg *BasicEngine) Run() error {
	eg.logger.Info(getBanner())
	registry.Validate()
	eg.logger.Info("scene service initialized successfully")
	eg.printContainersInfo()
	eg.logger.Info("starting scene engine...")
	if err := eg.Start(); err != nil {
		eg.logger.Error("start scene engine encounter an error, please fix error and restart")
		return errStartEngineFailed
	}
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	sig := <-quit
	eg.logger.Infof("received %v signal, shutting down...", sig)
	eg.Stop()
	eg.logger.Info("scene service stopped")
	return nil
}

func (eg *BasicEngine) Start() error {
	for _, setupable := range registry.Setupable.AcquireAll() {
		err := setupable.Setup()
		if err != nil {
			eg.logger.Errorf("setup %v error: %v", reflect.TypeOf(setupable), err)
			return err
		}
	}
	for _, container := range eg.containers {
		if err := container.Start(); err != nil {
			eg.logger.Errorf("start container %s error: %s", container.Name(), err)
			return err
		}
	}
	return nil
}

func (eg *BasicEngine) Stop() {
	ctx := context.Background()
	for _, container := range eg.containers {
		if err := container.Stop(ctx); err != nil {
			eg.logger.Errorf("stop container %s error: %s", container.Name(), err)
		}
	}
	for _, disposable := range registry.Disposable.AcquireAll() {
		err := disposable.Dispose()
		if err != nil {
			eg.logger.Warnf("dispose %v error: %v", reflect.TypeOf(disposable), err)
		}
	}
	return
}

func (eg *BasicEngine) ListContainers() []scene.ApplicationContainer {
	var containers []scene.ApplicationContainer
	for _, container := range eg.containers {
		containers = append(containers, container)
	}
	return containers
}

func (eg *BasicEngine) GetContainer(name string) scene.ApplicationContainer {
	return eg.containers[name]
}

func (eg *BasicEngine) AddContainer(container scene.ApplicationContainer) error {
	if _, exists := eg.containers[container.Name().Identifier()]; exists {
		panic(fmt.Sprintf("container %s already exists", container.Name()))
	}
	eg.logger.Infof("add container %s", container.Name())
	eg.containers[container.Name().Identifier()] = container
	return nil
}

//func (eg *BasicEngine) StopContainer(name string) error {
//	if container, exists := eg.containers[name]; exists {
//		eg.logger.Infof("stopping builder %s", container.Name())
//		return container.Stop(context.Background())
//	}
//	return errcode.AppContainerNotFound.WithDetailStr(name)
//}
//
//func (eg *BasicEngine) StartContainer(name string) error {
//	if container, exists := eg.containers[name]; exists {
//		eg.logger.Infof("starting builder %s", container.Name())
//		return container.Start()
//	}
//	return errcode.AppContainerNotFound.WithDetailStr(name)
//}
