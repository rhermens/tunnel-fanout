package keystore

import (
	"context"
	"log/slog"

	"github.com/google/go-github/v75/github"
	"github.com/spf13/viper"
)

type GitHubConfig struct {
	Organization string
	Token        string
}

func NewFromGithubOrganization() Keystore {
	var authorizedKeys []string
	config := GitHubConfig{
		Organization: viper.GetString("registry.ssh.github.organization"),
		Token:        viper.GetString("registry.ssh.github.token"),
	}

	slog.Info("Fetching authorized keys from GitHub organization", "organization", config.Organization)

	client := github.NewClient(nil).WithAuthToken(config.Token)
	members, _, err := client.Organizations.ListMembers(context.Background(), config.Organization, nil)
	if err != nil {
		panic(err)
	}

	for _, member := range members {
		keys, _, err := client.Users.ListKeys(context.Background(), member.GetLogin(), nil)
		if err != nil {
			panic(err)
		}

		for _, key := range keys {
			authorizedKeys = append(authorizedKeys, key.GetKey())
		}
	}

	return NewFromStrings(authorizedKeys)
}
