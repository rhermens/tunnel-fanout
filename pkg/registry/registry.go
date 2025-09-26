package registry

import (
	"log/slog"
	"net"
)

var Upstreams map[string]*TunnelUpstream = make(map[string]*TunnelUpstream)

func Listen(config *RegistryConfig) error {
	listener, err := net.Listen("tcp", net.JoinHostPort(config.Host, config.Port))
	if err != nil {
		slog.Error("Failed to start SSH server", "error", err)
		return err
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
		Upstreams[tu.RemoteAddr().String()] = tu

		tu.wg.Go(func() {
			err := tu.Wait()
			RemoveUpstream(tu, err)
		})
	}
}

func RemoveUpstream(tu *TunnelUpstream, reason error) {
	tu.Close()
	slog.Info("Connection closed", "remote", tu.SSHConn.RemoteAddr(), "local", tu.SSHConn.LocalAddr(), "reason", reason)
	delete(Upstreams, tu.RemoteAddr().String())
}
