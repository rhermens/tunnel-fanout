package cmd

import (
	"github.com/rhermens/tunnel-fanout/pkg/listener"
	"github.com/spf13/cobra"
)

func NewListenCmd() *cobra.Command {
	listenCmd := &cobra.Command{
		Use:   "listen",
		Short: "Listen to a tunnel",
		Run: func(cmd *cobra.Command, args []string) {
			lPort, err := cmd.Flags().GetInt("port")
			if err != nil {
				panic(err)
			}

			l := &listener.TunnelClient{
				LocalTargetPort: lPort,
			}

			l.Listen()
			defer l.Close()
		},
	}

	listenCmd.Flags().IntP("port", "p", 0, "Port to listen on")
	return listenCmd
}
