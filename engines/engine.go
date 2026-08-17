package engines

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	"github.com/rhine-tech/scene/utils/must"
)

var errStartEngineFailed = fmt.Errorf("failed to start engine")

type BasicEngine struct {
	logger      logger.ILogger
	loader      *scene.ModuleLoader
	definitions []scene.SceneDefinition
	containers  []scene.Scene
	byName      map[string]scene.Scene
	started     int
	running     bool
}

// NewEngine creates an Engine that owns module initialization, lifecycle, and
// Scene construction. No initialization runs until Start or Run is called.
func NewEngine(loader *scene.ModuleLoader, definitions ...scene.SceneDefinition) scene.Engine {
	if loader == nil {
		panic("scene engine: module loader is nil")
	}
	return &BasicEngine{
		logger:      logger.NoopLogger{},
		loader:      loader,
		definitions: append([]scene.SceneDefinition(nil), definitions...),
		byName:      make(map[string]scene.Scene),
	}
}

func (eg *BasicEngine) printContainersInfo() {
	containers := eg.ListContainers()
	info := make([]string, len(containers))
	for index, container := range containers {
		info[index] = formatContainerInfo(index, container)
	}
	eg.logger.Infof("successfully loaded %d containers. \n\n%s", len(containers), strings.Join(info, "\n"))
}

func (eg *BasicEngine) Run() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)
	if err := eg.Start(); err != nil {
		eg.logger.Errorf("start scene engine encounter an error, please fix error and restart: %v", err)
		return fmt.Errorf("%w: %w", errStartEngineFailed, err)
	}
	sig := <-quit
	eg.logger.Infof("received %v signal, shutting down...", sig)
	eg.Stop()
	eg.logger.Info("scene service stopped")
	return nil
}

func (eg *BasicEngine) Start() error {
	if eg.running {
		return nil
	}
	if err := eg.loader.Init(); err != nil {
		return err
	}
	eg.resolveLogger()
	eg.logger.Info(getBanner())
	eg.logger.Infof("App: %s", scene.AppName)
	eg.logger.Infof("Build: %s (%s) at %s",
		scene.AppBuildVersion,
		scene.AppBuildHash[:8],
		time.Unix(must.Must(strconv.ParseInt(scene.AppBuildTime, 10, 64)), 0).Format("2006-01-02 15:04:05"))
	eg.logger.Info("starting scene engine...")
	if err := eg.loader.Setup(); err != nil {
		eg.logger.Errorf("setup modules error: %v", err)
		return err
	}
	if err := eg.buildScenes(); err != nil {
		eg.logger.Errorf("build scenes error: %v", err)
		_ = eg.loader.TearDown()
		return err
	}
	eg.printContainersInfo()
	for _, container := range eg.containers {
		if err := container.Start(); err != nil {
			eg.logger.Errorf("start container %s error: %s", container.ImplName(), err)
			eg.stopContainers()
			_ = eg.loader.TearDown()
			return err
		}
		eg.started++
	}
	eg.running = true
	eg.logger.Info("scene service initialized successfully")
	return nil
}

func (eg *BasicEngine) Stop() {
	eg.stopContainers()
	if err := eg.loader.TearDown(); err != nil {
		eg.logger.Warnf("tear down modules error: %v", err)
	}
	eg.running = false
}

func (eg *BasicEngine) ListContainers() []scene.Scene {
	return append([]scene.Scene(nil), eg.containers...)
}

func (eg *BasicEngine) GetContainer(name string) scene.Scene {
	return eg.byName[name]
}

func (eg *BasicEngine) resolveLogger() {
	log, ok := registry.LookupIn[logger.ILogger](eg.loader.Scope())
	if !ok {
		log = logger.NoopLogger{}
	}
	eg.logger = log.WithPrefix("scene.engine")
}

func (eg *BasicEngine) buildScenes() error {
	containers := make([]scene.Scene, 0, len(eg.definitions))
	byName := make(map[string]scene.Scene, len(eg.definitions))
	for index, definition := range eg.definitions {
		container, err := definition.Build(eg.loader)
		if err != nil {
			return err
		}
		if container == nil {
			return fmt.Errorf("scene engine: scene %d is nil", index)
		}
		name := container.ImplName().Identifier()
		if _, exists := byName[name]; exists {
			return fmt.Errorf("scene engine: scene %s already exists", name)
		}
		byName[name] = container
		containers = append(containers, container)
	}
	eg.containers = containers
	eg.byName = byName
	return nil
}

func (eg *BasicEngine) stopContainers() {
	ctx := context.Background()
	for eg.started > 0 {
		eg.started--
		container := eg.containers[eg.started]
		if err := container.Stop(ctx); err != nil {
			eg.logger.Errorf("stop container %s error: %s", container.ImplName(), err)
		}
	}
}
