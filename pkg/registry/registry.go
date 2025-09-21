package registry

import (
	"log/slog"
	"net"
)

var Upstreams []*TunnelUpstream

func Listen(config *RegistryConfig) {
	listener, err := net.Listen("tcp", net.JoinHostPort(config.Host, config.Port))
	if err != nil {
		slog.Error("Failed to start SSH server", "error", err)
		panic(err)
	}
	slog.Info("SSH server listening on", "host", config.Host, "port", config.Port)
	defer listener.Close()

	for {
		var err error
		nConn, err := listener.Accept()
		if err != nil {
			slog.Error("Failed to accept connection", "error", err)
			continue
		}

		tu, err := NewUpstreamFromTCP(nConn, config)
		if err != nil {
			slog.Error("Failed to create tunnel upstream", "error", err)
			continue
		}

		slog.Info("Accepted connection", "remote", tu.SSHConn.RemoteAddr(), "local", tu.SSHConn.LocalAddr())
		Upstreams = append(Upstreams, tu)
	}
}
