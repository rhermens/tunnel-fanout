package main

import "github.com/rhermens/tunnel-fanout/pkg/cmd"

func main() {
	cmd := cmd.NewTunnelCmd()
	cmd.Execute()
}
