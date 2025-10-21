package registry

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/rhermens/tunnel-fanout/pkg/registry/keystore"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

type SshConfig struct {
	AuthorizedKeys keystore.Keystore
	HostKeyPath    string
	SshConfig      *ssh.ServerConfig
}

type RegistryConfig struct {
	Host string
	Port string
	Ssh  *SshConfig
}

func SetConfigDefaults() {
	viper.SetDefault("registry.host", "0.0.0.0")
	viper.SetDefault("registry.port", "7891")
	viper.SetDefault("registry.ssh.host_key_path", ".ssh/id_ed25519")
	viper.SetDefault("registry.ssh.authorized_keys", []string{})
	viper.SetDefault("registry.ssh.github.organization", nil)
	viper.SetDefault("registry.ssh.github.token", nil)
}

func NewRegistryConfig() *RegistryConfig {
	return &RegistryConfig{
		Host: viper.GetString("registry.host"),
		Port: viper.GetString("registry.port"),
		Ssh:  newSshConfig(),
	}
}

func newSshConfig() *SshConfig {
	authorizedKeys := keystore.NewFromYaml()
	if viper.IsSet("registry.ssh.github.organization") && viper.IsSet("registry.ssh.github.token") {
		authorizedKeys = keystore.MergeKeystores(authorizedKeys, keystore.NewFromGithubOrganization())
	}

	config := &SshConfig{
		AuthorizedKeys: authorizedKeys,
		HostKeyPath:    viper.GetString("registry.ssh.host_key_path"),
		SshConfig: &ssh.ServerConfig{
			NoClientAuth:      true,
			PublicKeyCallback: newPublicKeyCallback(authorizedKeys),
			NoClientAuthCallback: func(conn ssh.ConnMetadata) (*ssh.Permissions, error) {
				ip := conn.RemoteAddr().(*net.TCPAddr).IP
				if ip.IsLoopback() || ip.IsPrivate() {
					slog.Info("Allowing connection without authentication", "user", conn.User, "remote", conn.RemoteAddr())
					return &ssh.Permissions{
						Extensions: map[string]string{
							"no-auth": "true",
						},
					}, nil
				}

				slog.Warn("Denied connection without authentication", "user", conn.User(), "remote", conn.RemoteAddr().String(), "ip", ip.String())
				return nil, fmt.Errorf("authentication required for %q", conn.User())
			},
		},
	}

	config.SshConfig.AddHostKey(EnsureHostKey(config.HostKeyPath))
	return config
}

func newPublicKeyCallback(authorizedKeys map[string]bool) func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	return func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		if authorizedKeys[string(key.Marshal())] {
			return &ssh.Permissions{
				Extensions: map[string]string{
					"pubkey-fp": ssh.FingerprintSHA256(key),
				},
			}, nil
		}

		slog.Warn("Unauthorized public key", "user", conn.User(), "remote", conn.RemoteAddr(), "key", key.Type())
		return nil, fmt.Errorf("unauthorized public key for %q", conn.User())
	}
}
