package cmd

import (
	"github.com/rhine-tech/scene"
	"github.com/spf13/cobra"
)

type CmdApp interface {
	scene.Application
	Command(rootCmd *cobra.Command) error
}

type Container interface {
	scene.Named
	Execute() error
	ListAppNames() []string
	RootCommand() *cobra.Command
}
