package client

import (
	"os"

	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

type TunnelClientConfig struct {
	Registry    string
	TargetProto string
	TargetPort  int
	SshConfig   *ssh.ClientConfig
}

func NewTunnelClientConfig() *TunnelClientConfig {
	hostName, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	pk, err := os.ReadFile(viper.GetString("private_key_path"))
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
		TargetPort:  viper.GetInt("port"),
		TargetProto: "http",
		Registry:    viper.GetString("registry"),
		SshConfig:   clientConfig,
	}
}
