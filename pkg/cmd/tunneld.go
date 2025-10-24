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
		Short: "Start standalone tunnel daemon",
		Run: func(cmd *cobra.Command, args []string) {
			NewStandaloneCmd().Run(cmd, args)
		},
	}

	tunneldConfig()
	serveCmd.AddCommand(NewRegistryCmd())
	serveCmd.AddCommand(NewProxyCmd())
	serveCmd.AddCommand(NewStandaloneCmd())
	return serveCmd
}

func NewStandaloneCmd() *cobra.Command {
	standaloneCmd := &cobra.Command{
		Use:   "standalone",
		Short: "Start standalone tunneld server",
		Run: func(cmd *cobra.Command, args []string) {
			var wg sync.WaitGroup
			registry := registry.NewRegistry(registry.NewRegistryServerConfig())
			proxy := proxy.NewStandaloneHttpProxy(proxy.NewHttpServerConfig(), &registry)

			wg.Go(func() {
				err, _ := registry.Listen()
				if err != nil {
					slog.Error("Failed to start registry server", "error", err)
					os.Exit(1)
				}
			})

			wg.Go(func() {
				err := proxy.Listen()
				if err != nil {
					slog.Error("Failed to start http server", "error", err)
					os.Exit(1)
				}
			})

			wg.Wait()
		},
	}

	return standaloneCmd
}

func NewRegistryCmd() *cobra.Command {
	registryCmd := &cobra.Command{
		Use:   "registry",
		Short: "Start registry server",
		Run: func(cmd *cobra.Command, args []string) {
			registry := registry.NewRegistry(registry.NewRegistryServerConfig())
			_, err := registry.Listen()
			if err != nil {
				slog.Error("Failed to start registry server", "error", err)
				os.Exit(1)
			}
		},
	}

	return registryCmd
}

func NewProxyCmd() *cobra.Command {
	proxyCmd := &cobra.Command{
		Use:   "proxy",
		Short: "Start proxy server",
		Run: func(cmd *cobra.Command, args []string) {
			proxy := proxy.NewHttpProxy(proxy.NewHttpServerConfig())
			err := proxy.Listen()
			if err != nil {
				slog.Error("Failed to start http server", "error", err)
				os.Exit(1)
			}
		},
	}

	return proxyCmd
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
