package registry

import (
	"log/slog"
	"net"
	"sync"
)

type Registry struct {
	Config      *RegistryServerConfig
	Connections sync.Map
}

func NewRegistry(config *RegistryServerConfig) Registry {
	return Registry{
		Config: config,
	}
}

func (r *Registry) Listen() (*Registry, error) {
	listener, err := net.Listen("tcp", net.JoinHostPort(r.Config.Host, r.Config.Port))
	if err != nil {
		slog.Error("Failed to start registry server", "error", err)
		return nil, err
	}
	slog.Info("Registry server listening on", "host", r.Config.Host, "port", r.Config.Port)
	defer listener.Close()

	for {
		var err error
		nConn, err := listener.Accept()
		if err != nil {
			slog.Error("Failed to accept connection", "error", err)
			continue
		}

		err = r.AddConnection(nConn)
		if err != nil {
			slog.Error("Failed to create open connection", "error", err)
			continue
		}
	}
}

func (r *Registry) CloseConnection(c *Connection, reason error) {
	c.Close()
	slog.Info("Connection closed", "remote", c.RemoteAddr(), "local", c.LocalAddr(), "reason", reason)
	r.Connections.Delete(c.RemoteAddr().String())
}

func (r *Registry) FanoutBuffer(data []byte) {
	r.Connections.Range(func(key, value any) bool {
		connection := value.(*Connection)
		if connection.Type != Client {
			return true
		}

		go func() {
			if err := connection.ForwardBuffer(data); err != nil {
				slog.Error("Failed to send request", "error", err)
				r.CloseConnection(connection, err)
			}
		}()

		return true
	})
}
