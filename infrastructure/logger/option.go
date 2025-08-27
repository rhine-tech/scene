package logger

const (
	OptionCallerSkip = "caller-skip"
	OptionWithPrefix = "with-prefix"
)

type Option struct {
	Name  string
	Value interface{}
}

func WithCallerSkip(skip int) Option {
	return Option{
		Name:  OptionCallerSkip,
		Value: skip,
	}
}

func WithPrefix(prefix string) Option {
	return Option{
		Name:  OptionWithPrefix,
		Value: prefix,
	}
}
