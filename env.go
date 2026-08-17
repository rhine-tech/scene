package scene

import "os"

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
	EnvTest        = "test"
)

var Environment = EnvDevelopment

func init() {
	Environment = resolveEnvironment(Environment, os.Getenv("SCENE_ENV"))
}

func resolveEnvironment(defaultValue, override string) string {
	if override != "" {
		defaultValue = override
	}
	switch defaultValue {
	case EnvProduction, EnvDevelopment, EnvTest:
		return defaultValue
	default:
		return EnvDevelopment
	}
}
