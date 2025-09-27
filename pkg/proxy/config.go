package proxy

import "github.com/spf13/viper"

type HttpServerConfig struct {
	Host string
	Port string
}

func NewHttpServerConfig() *HttpServerConfig {
	return &HttpServerConfig{
		Host: viper.GetString("http.host"),
		Port: viper.GetString("http.port"),
	}
}

func SetConfigDefaults() {
	viper.SetDefault("http.host", "0.0.0.0")
	viper.SetDefault("http.port", "8000")
}
