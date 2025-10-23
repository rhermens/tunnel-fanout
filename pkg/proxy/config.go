package proxy

import (
	"github.com/rhermens/tunnel-fanout/pkg/registry"
	"github.com/spf13/viper"
)

type HttpServerConfig struct {
	Host           string
	Port           string
	Paths          []string
	RegistryConfig registry.RegistryClientConfig
}

func NewHttpServerConfig() *HttpServerConfig {
	cfg := &HttpServerConfig{
		Host:  viper.GetString("http.host"),
		Port:  viper.GetString("http.port"),
		Paths: make([]string, 0),
		RegistryConfig: registry.RegistryClientConfig{
			Address:    viper.GetString("http.registry.address"),
			SshKeyPath: viper.GetString("http.registry.ssh_key_path"),
			SshConfig:  registry.NewSshClientConfig(),
		},
	}
	viper.UnmarshalKey("http.paths", &cfg.Paths)

	if cfg.RegistryConfig.Address == "" {
		return cfg
	}

	cfg.RegistryConfig.AddSshAuth()
	return cfg
}

func SetConfigDefaults() {
	viper.SetDefault("http.host", "0.0.0.0")
	viper.SetDefault("http.port", "8000")
	viper.SetDefault("http.paths", []string{"/{path...}"})
	viper.SetDefault("http.registry.address", "localhost:7891")
}
