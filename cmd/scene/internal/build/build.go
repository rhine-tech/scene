package build

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
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
var outName string
var buildVersion string
var buildHash string
var env string
var staticBinary bool

func init() {
	// Build command flags
	CmdBuild.Flags().StringVar(&osTarget, "os", runtime.GOOS, "Target operating system (darwin, windows, linux, all)")
	CmdBuild.Flags().StringVar(&buildDir, "build-dir", "./dist", "Directory to place the build output")
	CmdBuild.Flags().StringVar(&outName, "out", "", "Name of the output executable")
	CmdBuild.Flags().StringVar(&buildVersion, "version", "v0.0.0", "Version of the application")
	CmdBuild.Flags().StringVar(&buildHash, "build-hash", "", "Git hash of the application, default is the current git hash")
	CmdBuild.Flags().StringVar(&env, "env", "production", "Environment of the application (development, production, test), default is production")
	CmdBuild.Flags().BoolVar(&staticBinary, "static", false, "Build a static binary (CGO_ENABLED=0)")
}

func build(cmd *cobra.Command, args []string) {
	packagePath := args[0] // The first (and only) argument is the package path
	if outName == "" {
		outName = filepath.Base(packagePath) + "_server"
	}

	// Handle building for different OS targets
	switch osTarget {
	case "darwin", "windows", "linux":
		executeBuild(osTarget, packagePath, buildDir, outName)
	case "all":
		executeBuild("darwin", packagePath, buildDir, outName+"-darwin")
		executeBuild("linux", packagePath, buildDir, outName+"-linux")
		executeBuild("windows", packagePath, buildDir, outName+"-windows")
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

	varName := func(name string) string {
		return "github.com/rhine-tech/scene" + "." + name
	}

	if buildHash == "" {
		if output, err := exec.Command("git", "rev-parse", "HEAD").Output(); err == nil {
			// check if is a valid git hash
			gitHash := regexp.MustCompile(`^[0-9a-f]{40}$`).FindString(strings.TrimSpace(string(output)))
			if gitHash != "" {
				buildHash = gitHash
			}
		} else {
			buildHash = "0000000000000000000000000000000000000000"
		}
	} else {
		// check if is a valid git hash
		gitHash := regexp.MustCompile(`^[0-9a-f]{40}$`).FindString(strings.TrimSpace(buildHash))
		if gitHash == "" {
			_, _ = fmt.Fprintf(os.Stderr, "Invalid git hash: %s\n", buildHash)
			os.Exit(1)
		}
	}

	ldflags := fmt.Sprintf("-ldflags=-X '%s=%d' -X '%s=%s' -X '%s=%s' -X '%s=%s'",
		varName("AppBuildTime"), time.Now().Unix(),
		varName("AppBuildHash"), buildHash,
		varName("AppBuildVersion"), buildVersion,
		varName("DEFAULT_ENV"), env,
	)

	appName := filepath.Base(packagePath)
	cmd := exec.Command("go", "build", "-o", outputPath, ldflags, packagePath)
	cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH=amd64")
	if staticBinary {
		cmd.Env = append(cmd.Env, "CGO_ENABLED=0")
	}
	output, err := cmd.CombinedOutput() // Capture both stdout and stderr
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[Scene Build] failed to build app '%s' for %s: %v\n\n%s\n", appName, goos, err, output)
		os.Exit(1)
	}
	_, _ = fmt.Fprintf(os.Stdout, "[Scene Build] successfully built app '%s' - %s (%s) for %s at %s\n",
		appName, buildVersion, buildHash[:8],
		goos, outputPath)
}
