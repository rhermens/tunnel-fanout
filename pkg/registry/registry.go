package registry

import (
	"log/slog"
	"net"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"
)

var TunnelClients []*TunnelClient

func Listen(addr string) {
	// authorizedKeys := []string{}

	serverConfig := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			return nil, nil
		},
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			return nil, nil
		},
	}

	pkBytes, err := os.ReadFile(".ssh/id_ed25519")
	if err != nil {
		slog.Error("Failed to load private key", "error", err)
		panic(err)
	}
	pk, err := ssh.ParsePrivateKey(pkBytes)
	if err != nil {
		slog.Error("Failed to parse private key", "error", err)
		panic(err)
	}
	serverConfig.AddHostKey(pk)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("Failed to start SSH server", "error", err)
		panic(err)
	}
	slog.Info("SSH server listening on", "addr", addr)
	defer listener.Close()

	var wg sync.WaitGroup
	defer wg.Done()
	for {
		var err error
		var tunnelClient TunnelClient
		nConn, err := listener.Accept()
		if err != nil {
			slog.Error("Failed to accept connection", "error", err)
			panic(err)
		}

		tunnelClient.SSHConn, tunnelClient.ChannelRequests, tunnelClient.Reqs, err = ssh.NewServerConn(nConn, serverConfig)
		if err != nil {
			slog.Error("Failed to handshake", "error", err)
			continue
		}
		wg.Go(func() {
			ssh.DiscardRequests(tunnelClient.Reqs)
		})

		slog.Info("Accepted connection", "remote", tunnelClient.SSHConn.RemoteAddr(), "local", tunnelClient.SSHConn.LocalAddr())
		TunnelClients = append(TunnelClients, &tunnelClient)

		for channel := range tunnelClient.ChannelRequests {
			c, reqs, err := channel.Accept()
			if err != nil {
				slog.Error("Could not accept channel", "error", err)
				continue
			}

			slog.Info("Accepted channel req", "remote", tunnelClient.SSHConn.RemoteAddr(), "local", tunnelClient.SSHConn.LocalAddr())
			tunnelClient.OpenChannels = append(tunnelClient.OpenChannels, OpenChannel{Channel: c, Requests: reqs})
		}
	}
}
