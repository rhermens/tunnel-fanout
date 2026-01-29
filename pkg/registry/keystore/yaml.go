package keystore

import (
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

type YamlKeystore struct {
	authorizedKeys map[string]bool
}

func NewYamlKeystore() *YamlKeystore {
	var authorizedKeys []string
	viper.UnmarshalKey("registry.ssh.authorized_keys", &authorizedKeys)

	return &YamlKeystore{
		authorizedKeys: AuthorizedKeysFromStrings(authorizedKeys),
	}
}

func (yk *YamlKeystore) ContainsKey(key ssh.PublicKey) bool {
	return yk.authorizedKeys[string(key.Marshal())]
}
