package cmd

import (
	"log/slog"
	"os"
	"strings"
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
			var wg sync.WaitGroup

			wg.Go(func() {
				err := proxy.Listen(proxy.NewHttpServerConfig())
				if err != nil {
					slog.Error("Failed to start http server", "error", err)
					os.Exit(1)
				}
			})
			wg.Go(func() {
				err := registry.Listen(registry.NewRegistryConfig())
				if err != nil {
					slog.Error("Failed to start registry server", "error", err)
					os.Exit(1)
				}
			})

			wg.Wait()
		},
	}

	tunneldConfig()
	return serveCmd
}

func tunneldConfig() {
	viper.SetConfigName("tunneld")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/tunneld/")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("tunneld")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	proxy.SetConfigDefaults()
	registry.SetConfigDefaults()

	err := viper.ReadInConfig()
	if err != nil {
		slog.Warn("No config file found", "error", err)
	}
}
