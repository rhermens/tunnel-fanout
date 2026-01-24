package main

import "github.com/rhermens/tunneld/pkg/cmd"

func main() {
	cmd := cmd.NewTunnelCmd()
	cmd.Execute()
}
