package tempfile

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/rhine-tech/scene/registry"
)

const (
	defaultTempDir     = "/tmp/scene"
	fallbackTempDir    = "./tempfile"
	configTempDirKey   = "scene.temp_dir"
	defaultTempPattern = "tmp-*"
)

var invalidSegmentRe = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

var (
	resolvedBaseDir string
	resolveOnce     sync.Once
)

func BaseDir() string {
	resolveOnce.Do(func() {
		resolvedBaseDir = resolveBaseDir()
	})
	return resolvedBaseDir
}

func ModuleDir(module string) string {
	seg := sanitizeSegment(module)
	if seg == "" {
		seg = "default"
	}
	return filepath.Join(BaseDir(), seg)
}

func EnsureModuleDir(module string) (string, error) {
	dir := ModuleDir(module)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func Create(module string, pattern string) (*os.File, error) {
	dir, err := EnsureModuleDir(module)
	if err != nil {
		return nil, err
	}
	pat := strings.TrimSpace(pattern)
	if pat == "" {
		pat = defaultTempPattern
	}
	return os.CreateTemp(dir, pat)
}

func SaveReader(module string, pattern string, src io.Reader) (string, func(), error) {
	tmp, err := Create(module, pattern)
	if err != nil {
		return "", nil, err
	}
	path := tmp.Name()
	cleanup := func() {
		_ = os.Remove(path)
	}
	if _, err := io.Copy(tmp, src); err != nil {
		_ = tmp.Close()
		cleanup()
		return "", nil, err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return "", nil, err
	}
	return path, cleanup, nil
}

func resolveBaseDir() string {
	candidate := defaultTempDir
	if registry.Config != nil {
		if configured := strings.TrimSpace(registry.Config.GetString(configTempDirKey)); configured != "" {
			candidate = configured
		}
	}
	candidate = filepath.Clean(candidate)
	if isUsableDir(candidate) {
		return candidate
	}
	fallback := filepath.Clean(fallbackTempDir)
	if isUsableDir(fallback) {
		return fallback
	}
	panic("tempfile: unable to create usable temp dir for both primary and fallback paths")
}

func isUsableDir(dir string) bool {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false
	}
	tmp, err := os.CreateTemp(dir, ".check-*")
	if err != nil {
		return false
	}
	name := tmp.Name()
	_ = tmp.Close()
	_ = os.Remove(name)
	return true
}

func sanitizeSegment(v string) string {
	s := strings.TrimSpace(v)
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "\\", "-")
	s = invalidSegmentRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, ".-")
	return s
}
