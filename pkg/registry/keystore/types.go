package keystore

import (
	"log/slog"
	"maps"

	"golang.org/x/crypto/ssh"
)

type Keystore = map[string]bool

func NewFromStrings(keys []string) Keystore {
	keystore := make(Keystore)
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

func MergeKeystores(keystores ...Keystore) Keystore {
	merged := make(Keystore)
	for _, ks := range keystores {
		maps.Copy(merged, ks)
	}

	return merged
}
