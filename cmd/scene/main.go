package main

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/cmd/scene/internal/build"
	"github.com/spf13/cobra"
	"log"
)

var rootCmd = &cobra.Command{
	Use:     "scene",
	Short:   "Scene: A lightweight microservice framework for Go.",
	Long:    `Scene: A lightweight microservice framework for Go.`,
	Version: scene.Version,
}

func init() {
	rootCmd.AddCommand(build.CmdBuild)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
