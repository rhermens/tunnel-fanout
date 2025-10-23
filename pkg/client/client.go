package client

import (
	"bufio"
	"bytes"
	"log/slog"
	"net"
	"net/http"

	"github.com/rhermens/tunnel-fanout/pkg/registry"
)

type TunnelClient struct {
	Config     *TunnelClientConfig
	Connection *registry.SshRegistryConnection
}

func NewTunnelClient(config *TunnelClientConfig) *TunnelClient {
	return &TunnelClient{
		Config: config,
	}
}

func (tc *TunnelClient) openRegistryConnection() error {
	var err error
	tc.Connection, err = registry.NewSshRegistryConnection(&tc.Config.RegistryConfig, registry.Client)
	if err != nil {
		return err
	}
	return nil
}

func (tc *TunnelClient) Close() {
	tc.Connection.Close()
}

func (tc *TunnelClient) ForwardRequests() {
	slog.Info("Forwarding requests", "port", tc.Config.TargetPort, "proto", tc.Config.TargetProto, "host", tc.Config.TargetHost)
	cl := http.Client{}
	defer tc.Close()

	for r := range tc.Connection.Requests {
		reader := bytes.NewReader(r.Payload)
		hReq, err := http.ReadRequest(bufio.NewReader(reader))
		if err != nil {
			slog.Error("Failed to read request", "error", err)
			continue
		}
		r.Reply(true, nil)

		hReq.URL.Scheme = tc.Config.TargetProto
		hReq.URL.Host = net.JoinHostPort(tc.Config.TargetHost, tc.Config.TargetPort)
		hReq.RequestURI = ""

		resp, err := cl.Do(hReq)
		if err != nil {
			slog.Error("Failed to do request", "error", err)
			continue
		}
		defer resp.Body.Close()
		slog.Info("Request forwarded", "url", hReq.URL, "status", resp.StatusCode)

		if r.WantReply {
			var buff bytes.Buffer
			resp.Write(&buff)

			tc.Connection.Channel.SendRequest("response", false, buff.Bytes())

			slog.Info("Response written back to proxy", "status", resp.StatusCode)
		}
	}
}

func (tc *TunnelClient) Listen() error {
	err := tc.openRegistryConnection()
	if err != nil {
		slog.Error("Failed to open registry connection", "error", err)
		return err
	}
	defer tc.Close()

	tc.ForwardRequests()
	return nil
}
