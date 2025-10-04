package client

import (
	"os"

	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

type TunnelClientConfig struct {
	Registry    string
	TargetProto string
	TargetHost  string
	TargetPort  string
	SshConfig   *ssh.ClientConfig
}

func NewTunnelClientConfig() *TunnelClientConfig {
	hostName, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	pk, err := os.ReadFile(viper.GetString("ssh_key_path"))
	if err != nil {
		panic(err)
	}
	signer, err := ssh.ParsePrivateKey(pk)
	if err != nil {
		panic(err)
	}

	clientConfig := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		User:            hostName,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}

	return &TunnelClientConfig{
		TargetHost:  viper.GetString("host"),
		TargetPort:  viper.GetString("port"),
		TargetProto: "http",
		Registry:    viper.GetString("registry"),
		SshConfig:   clientConfig,
	}
}
