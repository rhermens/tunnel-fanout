package registry

import (
	"log/slog"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
)

var TunnelClients []*TunnelUpstream

type RegistryConfig struct {
	SshConfig *ssh.ServerConfig
}

func newServerConfig() *RegistryConfig {
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
	return &RegistryConfig{SshConfig: serverConfig}
}

func Listen(addr string) {
	registryConfig := newServerConfig()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("Failed to start SSH server", "error", err)
		panic(err)
	}
	slog.Info("SSH server listening on", "addr", addr)
	defer listener.Close()

	for {
		var err error
		nConn, err := listener.Accept()
		if err != nil {
			slog.Error("Failed to accept connection", "error", err)
			continue
		}

		tu, err := NewUpstreamFromTCP(nConn, registryConfig)
		if err != nil {
			slog.Error("Failed to create tunnel upstream", "error", err)
			continue
		}

		slog.Info("Accepted connection", "remote", tu.SSHConn.RemoteAddr(), "local", tu.SSHConn.LocalAddr())
		TunnelClients = append(TunnelClients, tu)
	}
}
