package cfgur

import "strings"

func toUnderscoreKey(key string) string {
	return strings.ReplaceAll(strings.ToLower(key), ".", "_")
}
