package build

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var CmdBuild = &cobra.Command{
	Use:   "build [path]",
	Short: "Build the application",
	Args:  cobra.ExactArgs(1), // Requires exactly one argument: the path
	Run:   build,
}

// Global variables for command flags
var osTarget string
var buildDir string
var name string

func init() {
	// Build command flags
	CmdBuild.Flags().StringVar(&osTarget, "os", runtime.GOOS, "Target operating system (darwin, windows, linux, all)")
	CmdBuild.Flags().StringVar(&buildDir, "build-dir", "./dist", "Directory to place the build output")
	CmdBuild.Flags().StringVar(&name, "name", "", "Name of the output executable")
}

func build(cmd *cobra.Command, args []string) {
	packagePath := args[0] // The first (and only) argument is the package path
	if name == "" {
		name = filepath.Base(packagePath) + "_server"
	}

	// Handle building for different OS targets
	switch osTarget {
	case "darwin", "windows", "linux":
		executeBuild(osTarget, packagePath, buildDir, name)
	case "all":
		executeBuild("darwin", packagePath, buildDir, name+"-darwin")
		executeBuild("linux", packagePath, buildDir, name+"-linux")
		executeBuild("windows", packagePath, buildDir, name+"-windows")
	default:
		_, _ = fmt.Fprintf(os.Stderr, "Invalid or no --os specified. Use one of: darwin, windows, linux, all")
		os.Exit(1)
	}
}

func executeBuild(goos, packagePath, buildDir, outputName string) {
	if !filepath.IsAbs(packagePath) && !strings.HasPrefix(packagePath, "./") {
		packagePath = "./" + packagePath
	}
	outputPath := filepath.Join(buildDir, outputName)
	if goos == "windows" {
		outputPath += ".exe"
	}
	appName := filepath.Base(packagePath)
	cmd := exec.Command("go", "build", "-o", outputPath, packagePath)
	cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH=amd64")
	output, err := cmd.CombinedOutput() // Capture both stdout and stderr
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[Scene Build] failed to build app '%s' for %s: %v\n\n%s\n", appName, goos, err, output)
		os.Exit(1)
	}
	_, _ = fmt.Fprintf(os.Stdout, "[Scene Build] successfully built app '%s' for %s at %s\n", appName, goos, outputPath)
}
