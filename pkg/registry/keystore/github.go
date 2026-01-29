package keystore

import (
	"context"
	"log/slog"
	"sync"

	"github.com/google/go-github/v75/github"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

type GitHubConfig struct {
	Organization string
	Token        string
}

type GitHubKeystore struct {
	authorizedKeys map[string]bool
	config         GitHubConfig
	client         *github.Client
	mu             sync.Mutex
}

func NewGitHubKeystore() *GitHubKeystore {
	config := GitHubConfig{
		Organization: viper.GetString("registry.ssh.github.organization"),
		Token:        viper.GetString("registry.ssh.github.token"),
	}

	slog.Info("Fetching authorized keys from GitHub organization", "organization", config.Organization)

	client := github.NewClient(nil).WithAuthToken(config.Token)

	keystore := &GitHubKeystore{
		config: config,
		client: client,
	}
	keystore.Refresh()
	return keystore
}

func (k *GitHubKeystore) Refresh() {
	var authorizedKeys []string

	members, _, err := k.client.Organizations.ListMembers(context.Background(), k.config.Organization, nil)
	if err != nil {
		slog.Error("Failed to list organization members", "error", err)
		return
	}

	for _, member := range members {
		keys, _, err := k.client.Users.ListKeys(context.Background(), member.GetLogin(), nil)
		if err != nil {
			slog.Error("Failed to list user keys", "user", member.GetLogin(), "error", err)
			continue
		}

		for _, key := range keys {
			authorizedKeys = append(authorizedKeys, key.GetKey())
		}
	}

	k.mu.Lock()
	k.authorizedKeys = AuthorizedKeysFromStrings(authorizedKeys)
	k.mu.Unlock()
}

func (k *GitHubKeystore) ContainsKey(key ssh.PublicKey) bool {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.authorizedKeys[string(key.Marshal())]
}
