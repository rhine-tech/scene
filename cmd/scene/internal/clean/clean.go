package clean

import (
	"github.com/spf13/cobra"
	"log"
	"os"
)

var CmdClean = &cobra.Command{
	Use:   "clean",
	Short: "Clean up the build directories",
	Run:   clean,
}

func clean(cmd *cobra.Command, args []string) {
	err := os.RemoveAll("./dist")
	if err != nil {
		log.Printf("Failed to clean the build directory: %v\n", err)
		os.Exit(1)
	}
	log.Println("Build directory cleaned.")
}
