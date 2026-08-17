package cmd

import (
	"fmt"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	"github.com/spf13/cobra"
)

type cmdContainer struct {
	root    *cobra.Command
	apps    []CmdApp
	log     logger.ILogger
	loader  *scene.ModuleLoader
	built   bool
	options []RootOption
}

// NewRootContainer creates a command container backed by loader.
func NewRootContainer(root *cobra.Command, loader *scene.ModuleLoader, options ...RootOption) Container {
	if root == nil {
		panic("scene cmd: nil root command")
	}
	if loader == nil {
		panic("scene cmd: module loader is nil")
	}
	return &cmdContainer{
		root:    root,
		log:     logger.NoopLogger{},
		loader:  loader,
		options: options,
	}
}

// NewContainer creates a root command backed by loader.
func NewContainer(use, short string, loader *scene.ModuleLoader, options ...RootOption) Container {
	return NewRootContainer(&cobra.Command{
		Use:          use,
		Short:        short,
		SilenceUsage: true,
	}, loader, options...)
}

func (c *cmdContainer) ImplName() scene.ImplName {
	return scene.NewSceneImplNameNoVer("cmd", "Scene")
}

func (c *cmdContainer) build() error {
	if c.built {
		return nil
	}
	for _, opt := range c.options {
		if err := opt(c.root); err != nil {
			return err
		}
	}
	for _, app := range c.apps {
		if err := app.Command(c.root); err != nil {
			return fmt.Errorf("scene cmd: app %s register failed: %w", app.Name(), err)
		}
		c.log.Infof("registered command app %s", app.Name())
	}
	c.built = true
	return nil
}

func (c *cmdContainer) prepare() error {
	if c.built {
		return nil
	}
	if err := c.loader.Init(); err != nil {
		return err
	}
	c.apps = scene.Applications[CmdApp](c.loader.Applications())
	if log, exists := registry.LookupIn[logger.ILogger](c.loader.Scope()); exists {
		c.log = log.WithPrefix(c.ImplName().Identifier())
	}
	return c.build()
}

func (c *cmdContainer) Execute() error {
	if err := c.prepare(); err != nil {
		_ = c.loader.TearDown()
		return err
	}
	if err := c.loader.Setup(); err != nil {
		c.log.Errorf("setup modules error: %v", err)
		return err
	}
	defer func() {
		if err := c.loader.TearDown(); err != nil {
			c.log.Warnf("tear down modules error: %v", err)
		}
	}()
	c.log.Info("scene service initialized successfully")
	c.log.Infof("loaded %d command apps", len(c.apps))
	return c.root.Execute()
}

func (c *cmdContainer) RootCommand() *cobra.Command {
	if err := c.prepare(); err != nil {
		panic(err)
	}
	return c.root
}

func (c *cmdContainer) ListAppNames() []string {
	names := make([]string, 0, len(c.apps))
	for _, app := range c.apps {
		names = append(names, app.Name().Identifier())
	}
	return names
}
