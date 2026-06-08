package config

// TagName is the tag name used in config struct.
// The Tag contains two value. "config_key,default=def_value(optional)"
const TagName = "scfg"

type DefaultBehavior int

const (
	DefaultBehaviorUseTag DefaultBehavior = iota
	DefaultBehaviorZeroValue
)

type UnmarshalOptions struct {
	TagName         string
	Prefix          string
	DefaultBehavior DefaultBehavior
}

type UnmarshalOption func(*UnmarshalOptions)

type ConfigProvider interface {
	Init() error
	GetString(key string) string
	GetInt(key string) int64
	GetBool(key string) bool
	GetStringE(key string) (string, bool)
	GetIntE(key string) (int64, bool)
	GetBoolE(key string) (bool, bool)
}

type ConfigUnmarshaler interface {
	ConfigProvider
	Unmarshal(val interface{}) error
	UnmarshalWithPrefix(prefix string, val interface{}) error
}

type IConfig interface {
	ConfigUnmarshaler
	ConfigProviderWithDefault
}

func DefaultUnmarshalOptions() UnmarshalOptions {
	return UnmarshalOptions{
		TagName:         TagName,
		DefaultBehavior: DefaultBehaviorUseTag,
	}
}

func WithTagName(tagName string) UnmarshalOption {
	return func(options *UnmarshalOptions) {
		options.TagName = tagName
	}
}

func WithPrefix(prefix string) UnmarshalOption {
	return func(options *UnmarshalOptions) {
		options.Prefix = prefix
	}
}

func WithDefaultBehavior(behavior DefaultBehavior) UnmarshalOption {
	return func(options *UnmarshalOptions) {
		options.DefaultBehavior = behavior
	}
}
