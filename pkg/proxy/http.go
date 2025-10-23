package proxy

import (
	"bytes"
	"log/slog"
	"net"
	"net/http"

	"github.com/rhermens/tunnel-fanout/pkg/registry"
)

type RegistryConnection interface {
	Forward(data []byte) (int, error)
}

type SshRegistry struct {
	Connection *registry.SshRegistryConnection
}

func (c *SshRegistry) Forward(data []byte) (int, error) {
	_, err := c.Connection.Channel.SendRequest(string(registry.Forward), false, data)
	return len(data), err
}

type InMemoryRegistryConnection struct {
	Registry *registry.Registry
}

func (c InMemoryRegistryConnection) Forward(data []byte) (int, error) {
	c.Registry.FanoutBuffer(data)
	return len(data), nil
}

type HttpProxy struct {
	mux                *http.ServeMux
	Config             *HttpServerConfig
	RegistryConnection RegistryConnection
}

func NewStandaloneHttpProxy(config *HttpServerConfig, registry *registry.Registry) HttpProxy {
	proxy := HttpProxy{
		mux:    http.NewServeMux(),
		Config: config,
		RegistryConnection: InMemoryRegistryConnection{
			Registry: registry,
		},
	}

	return proxy
}

func NewHttpProxy(config *HttpServerConfig) HttpProxy {
	connection, err := NewSshRegistry(&config.RegistryConfig)
	if err != nil {
		slog.Error("Failed to create SSH registry connection", "error", err)
		panic(err)
	}

	proxy := HttpProxy{
		mux:                http.NewServeMux(),
		Config:             config,
		RegistryConnection: connection,
	}

	return proxy
}

func (p *HttpProxy) Listen() error {
	for _, handlePath := range p.Config.Paths {
		p.mux.HandleFunc(handlePath, p.ForwardHandler)
		slog.Info("Registered handler", "path", handlePath)
	}

	slog.Info("Starting http server", "host", p.Config.Host, "port", p.Config.Port)
	return http.ListenAndServe(net.JoinHostPort(p.Config.Host, p.Config.Port), p.mux)
}

func (p *HttpProxy) ForwardHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received request", "method", r.Method, "url", r.URL.String())
	var buffer bytes.Buffer
	r.Write(&buffer)

	_, err := p.RegistryConnection.Forward(buffer.Bytes())
	if err != nil {
		slog.Error("Failed to write to registry connection", "error", err)
	}

	w.Write([]byte("OK"))
}

func NewSshRegistry(config *registry.RegistryClientConfig) (*SshRegistry, error) {
	var err error
	conn, err := registry.NewSshRegistryConnection(config, registry.Proxy)

	if err != nil {
		return nil, err
	}

	return &SshRegistry{
		Connection: conn,
	}, nil
}
