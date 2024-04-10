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

func init() {
	switch os.Getenv("SCENE_ENV") {
	case "production":
		env = EnvProduction
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
