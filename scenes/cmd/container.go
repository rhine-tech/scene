package cmd

import (
	"fmt"
	"reflect"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	"github.com/spf13/cobra"
)

type cmdContainer struct {
	root    *cobra.Command
	apps    []CmdApp
	log     logger.ILogger
	built   bool
	options []RootOption
}

func NewRootContainer(root *cobra.Command, apps []CmdApp, options ...RootOption) Container {
	if root == nil {
		panic("scene cmd: nil root command")
	}
	return &cmdContainer{
		root:    root,
		apps:    apps,
		log:     registry.Logger.WithPrefix((&cmdContainer{}).ImplName().Identifier()),
		options: options,
	}
}

func NewContainer(use, short string, apps []CmdApp, options ...RootOption) Container {
	return NewRootContainer(&cobra.Command{
		Use:          use,
		Short:        short,
		SilenceUsage: true,
	}, apps, options...)
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

func (c *cmdContainer) Execute() error {
	if err := c.build(); err != nil {
		return err
	}
	registry.Validate()
	c.log.Info("scene service initialized successfully")
	c.log.Infof("loaded %d command apps", len(c.apps))
	for _, setupable := range registry.Setupable.AcquireAll() {
		if err := setupable.Setup(); err != nil {
			c.log.Errorf("setup %v error: %v", reflect.TypeOf(setupable), err)
			return err
		}
	}
	defer func() {
		for _, disposable := range registry.Disposable.AcquireAll() {
			if err := disposable.Dispose(); err != nil {
				c.log.Warnf("dispose %v error: %v", reflect.TypeOf(disposable), err)
			}
		}
	}()
	return c.root.Execute()
}

func (c *cmdContainer) RootCommand() *cobra.Command {
	if err := c.build(); err != nil {
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
