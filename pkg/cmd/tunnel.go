package cmd

import (
	"log/slog"
	"os"

	"github.com/rhermens/tunnel-fanout/pkg/listener"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewTunnelCmd() *cobra.Command {
	listenCmd := &cobra.Command{
		Use:   "tunnel",
		Short: "Listen to a tunnel",
		Run: func(cmd *cobra.Command, args []string) {
			config := listener.NewTunnelClientConfig()
			l := listener.NewTunnelClient(config)
			l.Listen()
			defer l.Close()
		},
	}

	viper.SetConfigName("tunnel")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(os.Getenv("HOME") + "/.config/tunnel/")
	viper.AddConfigPath(".")

	listenCmd.Flags().IntP("port", "p", 8080, "Port to forward to")
	listenCmd.Flags().StringP("registry", "r", ":7891", "Registry address")
	viper.BindPFlag("port", listenCmd.Flags().Lookup("port"))
	viper.BindPFlag("registry", listenCmd.Flags().Lookup("registry"))

	err := viper.ReadInConfig()
	if err != nil {
		slog.Warn("No config file found", "error", err)
	}

	return listenCmd
}
