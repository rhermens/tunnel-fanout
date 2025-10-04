package cmd

import (
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/rhermens/tunnel-fanout/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewTunnelCmd() *cobra.Command {
	tunnelCmd := &cobra.Command{
		Use:   "tunnel",
		Short: "Listen to a tunnel",
		Run: func(cmd *cobra.Command, args []string) {
			config := client.NewTunnelClientConfig()
			l := client.NewTunnelClient(config)
			err := l.Listen()
			if err != nil {
				slog.Error("Failed to listen to tunnel", "error", err)
				os.Exit(1)
			}

			defer l.Close()
		},
	}

	viper.SetConfigName("tunnel")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path.Join(os.Getenv("HOME"), ".config/tunnel"))
	viper.SetEnvPrefix("tunnel")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AddConfigPath(".")

	tunnelCmd.Flags().String("host", "localhost", "Host to forward to")
	tunnelCmd.Flags().StringP("port", "p", "8080", "Port to forward to")
	tunnelCmd.Flags().StringP("registry", "r", ":7891", "Registry address")
	tunnelCmd.Flags().StringP("ssh-key-path", "k", path.Join(os.Getenv("HOME"), ".ssh/id_ed25519"), "Public key path")
	viper.BindPFlag("host", tunnelCmd.Flags().Lookup("host"))
	viper.BindPFlag("port", tunnelCmd.Flags().Lookup("port"))
	viper.BindPFlag("registry", tunnelCmd.Flags().Lookup("registry"))
	viper.BindPFlag("ssh_key_path", tunnelCmd.Flags().Lookup("ssh-key-path"))

	err := viper.ReadInConfig()
	if err != nil {
		slog.Warn("No config file found", "error", err)
	}

	return tunnelCmd
}
