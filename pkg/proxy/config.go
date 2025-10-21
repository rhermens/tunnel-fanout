package proxy

import (
	"fmt"
	"net"
	"os"

	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

type RegistryConnectionConfig struct {
	Host      string
	Port      string
	SshConfig ssh.ClientConfig
}

type HttpServerConfig struct {
	Host  string
	Port  string
	Paths []string
}

func NewHttpServerConfig() *HttpServerConfig {
	var paths []string
	viper.UnmarshalKey("http.paths", &paths)

	return &HttpServerConfig{
		Host:  viper.GetString("http.host"),
		Port:  viper.GetString("http.port"),
		Paths: paths,
	}
}

func NewRegistryConnectionConfig() *RegistryConnectionConfig {
	cfg := &RegistryConnectionConfig{
		Host:      viper.GetString("http.registry.host"),
		Port:      viper.GetString("http.registry.port"),
		SshConfig: NewSshClientConfig(),
	}

	ips, err := net.LookupIP(cfg.Host)
	if err != nil || len(ips) == 0 {
		panic(fmt.Sprintf("Failed to resolve registry host: %s", cfg.Host))
	}

	ip := ips[0]
	if !ip.IsPrivate() && !ip.IsLoopback() {
		AddSshAuth(&cfg.SshConfig)
	}

	return cfg
}

func NewSshClientConfig() ssh.ClientConfig {
	hostName, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	return ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		User:            hostName,
		Auth:            []ssh.AuthMethod{},
	}
}

func AddSshAuth(cfg *ssh.ClientConfig) {
	pk, err := os.ReadFile(viper.GetString("http.registry.ssh_key_path"))
	if err != nil {
		panic(err)
	}
	signer, err := ssh.ParsePrivateKey(pk)
	if err != nil {
		panic(err)
	}

	cfg.Auth = []ssh.AuthMethod{
		ssh.PublicKeys(signer),
	}
}

func SetConfigDefaults() {
	viper.SetDefault("http.host", "0.0.0.0")
	viper.SetDefault("http.port", "8000")
	viper.SetDefault("http.paths", []string{"/{path...}"})
	viper.SetDefault("http.registry.host", "localhost")
	viper.SetDefault("http.registry.port", "7891")
}
