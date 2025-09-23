package registry

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

type SshConfig struct {
	AuthorizedKeys map[string]bool
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
	viper.SetDefault("registry.ssh.authorized_keys", []string{""})
}

func NewRegistryConfig() *RegistryConfig {
	return &RegistryConfig{
		Host: viper.GetString("registry.host"),
		Port: viper.GetString("registry.port"),
		Ssh:  newSshConfig(),
	}
}

func newSshConfig() *SshConfig {
	authorizedKeys := parseAuthorizedKeys(viper.GetStringSlice("registry.ssh.authorized_keys"))
	config := &SshConfig{
		AuthorizedKeys: authorizedKeys,
		HostKeyPath:    viper.GetString("registry.ssh.host_key_path"),
		SshConfig: &ssh.ServerConfig{
			PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
				if authorizedKeys[string(key.Marshal())] {
					return &ssh.Permissions{
						Extensions: map[string]string{
							"pubkey-fp": ssh.FingerprintSHA256(key),
						},
					}, nil
				}

				slog.Warn("Unauthorized public key", "user", conn.User(), "remote", conn.RemoteAddr(), "key", key.Type())
				return nil, fmt.Errorf("unauthorized public key for %q", conn.User())
			},
		},
	}

	pkBytes, err := os.ReadFile(config.HostKeyPath)
	if err != nil {
		slog.Error("Failed to load host key", "error", err, "path", config.HostKeyPath)
		panic(err)
	}
	pk, err := ssh.ParsePrivateKey(pkBytes)
	if err != nil {
		slog.Error("Failed to parse host key", "error", err)
		panic(err)
	}
	config.SshConfig.AddHostKey(pk)
	return config
}

func parseAuthorizedKeys(authorizedKeys []string) map[string]bool {
	authorizedKeysMap := map[string]bool{}

	for _, key := range authorizedKeys {
		if key == "" {
			continue
		}

		parsedKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
		if err != nil {
			slog.Warn("Failed to parse authorized key", "error", err)
			continue
		}

		authorizedKeysMap[string(parsedKey.Marshal())] = true
	}

	return authorizedKeysMap
}
