package client

import (
	"bufio"
	"bytes"
	"log/slog"
	"net"
	"net/http"

	"golang.org/x/crypto/ssh"
)

type TunnelClient struct {
	Config         *TunnelClientConfig
	RegistryClient *ssh.Client
	Channel        ssh.Channel
	Requests       <-chan *ssh.Request
}

func NewTunnelClient(config *TunnelClientConfig) *TunnelClient {
	return &TunnelClient{
		Config: config,
	}
}

func (tc *TunnelClient) openRegistryConnection() error {
	var err error
	tc.RegistryClient, err = ssh.Dial("tcp", tc.Config.Registry, tc.Config.SshConfig)
	if err != nil {
		return err
	}

	tc.Channel, tc.Requests, err = tc.RegistryClient.OpenChannel("upstream", []byte{})
	if err != nil {
		return err
	}

	slog.Info("Opened channel to registry", "remote", tc.Config.Registry)
	return nil
}

func (tc *TunnelClient) Close() {
	tc.RegistryClient.Close()
}

func (tc *TunnelClient) ForwardRequests() {
	slog.Info("Forwarding requests", "port", tc.Config.TargetPort, "proto", tc.Config.TargetProto, "host", tc.Config.TargetHost)
	cl := http.Client{}
	defer tc.Close()

	for r := range tc.Requests {
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

			tc.Channel.SendRequest("response", false, buff.Bytes())

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
