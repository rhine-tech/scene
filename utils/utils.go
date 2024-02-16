package utils

import (
	"github.com/spf13/viper"
	"net"
	"strconv"
	"strings"
)

func CreateViper(path string) *viper.Viper {
	v := viper.New()
	v.SetConfigFile(path)
	_ = v.ReadInConfig()
	return v
}

func IsValidAddress(address string) bool {
	if len(address) == 0 {
		return false
	}
	parts := strings.Split(address, ":")
	if len(parts) != 2 {
		return false
	}
	ip, portStr := parts[0], parts[1]

	if ip != "" && net.ParseIP(ip) == nil {
		return false
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return false
	}
	return true
}

func PrettyAddress(address string) string {
	if address[0] == ':' {
		address = "localhost" + address
	}
	return address
}
