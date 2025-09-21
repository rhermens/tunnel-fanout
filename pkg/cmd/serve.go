package cmd

import (
	"sync"

	"github.com/rhermens/tunnel-fanout/pkg/registry"
	"github.com/rhermens/tunnel-fanout/pkg/server"
	"github.com/spf13/cobra"
)

func NewServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the tunnel server",
		Run: func(cmd *cobra.Command, args []string) {
			var wg sync.WaitGroup

			wg.Add(1)
			go server.Listen(":9000")

			wg.Add(1)
			go registry.Listen(":8000")

			wg.Wait()
		},
	}
}
