package keystore

import (
	"log/slog"

	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

type Keystore interface {
	ContainsKey(key ssh.PublicKey) bool
}

type Keystores []Keystore

func AuthorizedKeysFromStrings(keys []string) map[string]bool {
	keystore := make(map[string]bool)
	for _, key := range keys {
		parsedKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
		if err != nil {
			slog.Warn("Failed to parse authorized key", "error", err)
			continue
		}

		slog.Debug("Parsed authorized key", "key", key)
		keystore[string(parsedKey.Marshal())] = true
	}

	return keystore
}

func LoadKeystores() Keystores {
	keystores := Keystores{}
	keystores = append(keystores, NewYamlKeystore())

	if viper.IsSet("registry.ssh.github.organization") && viper.IsSet("registry.ssh.github.token") {
		keystores = append(keystores, NewGitHubKeystore())
	}

	return keystores
}

func (ks Keystores) ContainsKey(key ssh.PublicKey) bool {
	for _, k := range ks {
		if k.ContainsKey(key) {
			return true
		}
	}

	return false
}
