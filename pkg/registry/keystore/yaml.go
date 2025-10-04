package keystore

import (
	"github.com/spf13/viper"
)

func NewFromYaml() Keystore {
	var authorizedKeys []string
	viper.UnmarshalKey("registry.ssh.authorized_keys", &authorizedKeys)

	return NewFromStrings(authorizedKeys)
}
