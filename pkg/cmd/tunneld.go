package cmd

import (
	"log/slog"
	"sync"

	"github.com/rhermens/tunnel-fanout/pkg/proxy"
	"github.com/rhermens/tunnel-fanout/pkg/registry"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewTunneldCmd() *cobra.Command {
	serveCmd := &cobra.Command{
		Use:   "tunneld",
		Short: "Start the tunnel daemon",
		Run: func(cmd *cobra.Command, args []string) {
			httpConf := viper.Sub("http")
			registryConf := viper.Sub("registry")

			var wg sync.WaitGroup

			wg.Add(1)
			go proxy.Listen(proxy.NewHttpServerConfig(httpConf))

			wg.Add(1)
			go registry.Listen(registry.NewRegistryConfig(registryConf))

			wg.Wait()
		},
	}

	serveCmd.Flags().String("http-host", ":8000", "HTTP Server host")
	serveCmd.Flags().String("registry-host", ":7891", "Tunnel Registry host")

	tunneldConfig()

	return serveCmd
}

func tunneldConfig() {
	viper.SetConfigName("tunneld")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/tunneld/")
	viper.AddConfigPath(".")
	proxy.SetConfigDefaults()
	registry.SetConfigDefaults()

	err := viper.ReadInConfig()
	if err != nil {
		slog.Warn("No config file found", "error", err)
	}
}
