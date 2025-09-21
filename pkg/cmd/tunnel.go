package cmd

import (
	"log/slog"
	"os"

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
			l.Listen()
			defer l.Close()
		},
	}

	viper.SetConfigName("tunnel")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(os.Getenv("HOME") + "/.config/tunnel/")
	viper.AddConfigPath(".")

	tunnelCmd.Flags().IntP("port", "p", 8080, "Port to forward to")
	tunnelCmd.Flags().StringP("registry", "r", ":7891", "Registry address")
	viper.BindPFlag("port", tunnelCmd.Flags().Lookup("port"))
	viper.BindPFlag("registry", tunnelCmd.Flags().Lookup("registry"))

	err := viper.ReadInConfig()
	if err != nil {
		slog.Warn("No config file found", "error", err)
	}

	return tunnelCmd
}
