package scene

import "os"

type Environment uint8

const (
	EnvDevelopment Environment = iota
	EnvProduction
	EnvTest
)

func (e Environment) String() string {
	switch e {
	case EnvDevelopment:
		return "development"
	case EnvProduction:
		return "production"
	case EnvTest:
		return "test"
	default:
		return "unknown"
	}
}

var env Environment = EnvDevelopment

// DEFAULT_ENV is the environment variable that can be set to "production", "test" or "development"
// this is the default
var DEFAULT_ENV string = "development"

func init() {
	// override the default environment setting
	envStr := DEFAULT_ENV
	if os.Getenv("SCENE_ENV") != "" {
		envStr = os.Getenv("SCENE_ENV")
	}
	switch envStr {
	case "production":
		env = EnvProduction
	case "development":
		env = EnvDevelopment
	case "test":
		env = EnvTest
	default:
		env = EnvDevelopment
	}
}

func SetEnvironment(e Environment) {
	env = e
}

func GetEnvironment() Environment {
	return env
}
