package registry

import (
	"log/slog"
	"os"

	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

type RegistryConfig struct {
	Host      string
	Port      string
	SshConfig *ssh.ServerConfig
}

func SetConfigDefaults() {
	viper.SetDefault("registry.host", "0.0.0.0")
	viper.SetDefault("registry.port", "7891")
}

func NewRegistryConfig(v *viper.Viper) *RegistryConfig {
	return &RegistryConfig{
		Host:      v.GetString("host"),
		Port:      v.GetString("port"),
		SshConfig: newSshConfig(),
	}
}

func newSshConfig() *ssh.ServerConfig {
	// authorizedKeys := []string{}
	serverConfig := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			return nil, nil
		},
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			return nil, nil
		},
	}

	pkBytes, err := os.ReadFile(".ssh/id_ed25519")
	if err != nil {
		slog.Error("Failed to load private key", "error", err)
		panic(err)
	}
	pk, err := ssh.ParsePrivateKey(pkBytes)
	if err != nil {
		slog.Error("Failed to parse private key", "error", err)
		panic(err)
	}
	serverConfig.AddHostKey(pk)
	return serverConfig
}
