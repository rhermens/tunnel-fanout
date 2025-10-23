package client

import (
	"github.com/rhermens/tunnel-fanout/pkg/registry"
	"github.com/spf13/viper"
)

type TunnelClientConfig struct {
	TargetProto    string
	TargetHost     string
	TargetPort     string
	RegistryConfig registry.RegistryClientConfig
}

func NewTunnelClientConfig() *TunnelClientConfig {
	cfg := &TunnelClientConfig{
		TargetHost:  viper.GetString("host"),
		TargetPort:  viper.GetString("port"),
		TargetProto: "http",
		RegistryConfig: registry.RegistryClientConfig{
			Address:    viper.GetString("registry"),
			SshKeyPath: viper.GetString("ssh_key_path"),
			SshConfig:  registry.NewSshClientConfig(),
		},
	}

	cfg.RegistryConfig.AddSshAuth()
	return cfg
}
