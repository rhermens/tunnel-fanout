package main

import (
	"github.com/rhermens/tunnel-fanout/pkg/cmd"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "tunnel",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func main() {
	rootCmd.AddCommand(cmd.NewListenCmd())
	rootCmd.AddCommand(cmd.NewServeCmd())
	rootCmd.Execute()
}
