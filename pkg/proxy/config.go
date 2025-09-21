package proxy

import "github.com/spf13/viper"

type HttpServerConfig struct {
	Host string
	Port string
}

func NewHttpServerConfig(v *viper.Viper) *HttpServerConfig {
	return &HttpServerConfig{
		Host: v.GetString("host"),
		Port: v.GetString("port"),
	}
}

func SetConfigDefaults() {
	viper.SetDefault("http.host", "0.0.0.0")
	viper.SetDefault("http.port", "8000")
}
