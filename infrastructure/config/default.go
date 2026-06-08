package config

type ConfigProviderWithDefault interface {
	ConfigProvider
	GetStringOrDefault(key, defaultValue string) string
	GetIntOrDefault(key string, defaultValue int64) int64
	GetBoolOrDefault(key string, defaultValue bool) bool
}

func GetStringOrDefault(provider ConfigProvider, key, defaultValue string) string {
	if provider == nil {
		return defaultValue
	}
	if val, ok := provider.GetStringE(key); ok {
		return val
	}
	return defaultValue
}

func GetIntOrDefault(provider ConfigProvider, key string, defaultValue int64) int64 {
	if provider == nil {
		return defaultValue
	}
	if val, ok := provider.GetIntE(key); ok {
		return val
	}
	return defaultValue
}

func GetBoolOrDefault(provider ConfigProvider, key string, defaultValue bool) bool {
	if provider == nil {
		return defaultValue
	}
	if val, ok := provider.GetBoolE(key); ok {
		return val
	}
	return defaultValue
}

func (m *providerMarshaller) GetStringOrDefault(key, defaultValue string) string {
	return GetStringOrDefault(m.ConfigProvider, key, defaultValue)
}

func (m *providerMarshaller) GetIntOrDefault(key string, defaultValue int64) int64 {
	return GetIntOrDefault(m.ConfigProvider, key, defaultValue)
}

func (m *providerMarshaller) GetBoolOrDefault(key string, defaultValue bool) bool {
	return GetBoolOrDefault(m.ConfigProvider, key, defaultValue)
}
