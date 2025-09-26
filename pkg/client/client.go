package client

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

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

func (tc *TunnelClient) openRegistryConnection() {
	var err error
	tc.RegistryClient, err = ssh.Dial("tcp", tc.Config.Registry, tc.Config.SshConfig)
	if err != nil {
		panic(err)
	}

	tc.Channel, tc.Requests, err = tc.RegistryClient.OpenChannel("upstream", []byte{})
	if err != nil {
		panic(err)
	}

	slog.Info("Opened channel to registry", "remote", tc.Config.Registry)
}

func (tc *TunnelClient) Close() {
	tc.RegistryClient.Close()
}

func (tc *TunnelClient) ForwardRequests() {
	slog.Info("Forwarding requests", "port", tc.Config.TargetPort, "proto", tc.Config.TargetProto)
	cl := http.Client{}
	defer tc.Close()

	for r := range tc.Requests {
		reader := bytes.NewReader(r.Payload)
		hReq, err := http.ReadRequest(bufio.NewReader(reader))
		if err != nil {
			slog.Error("Failed to read request", "error", err)
			continue
		}

		u, err := url.Parse(fmt.Sprintf("%s://localhost:%d%s", tc.Config.TargetProto, tc.Config.TargetPort, hReq.URL.Path))
		hReq.URL = u
		hReq.RequestURI = ""

		resp, err := cl.Do(hReq)
		if err != nil {
			slog.Error("Failed to do request", "error", err)
			continue
		}
		defer resp.Body.Close()
		slog.Info("Request forwarded", "path", hReq.URL.Path, "status", resp.StatusCode)

		if r.WantReply {
			var buff bytes.Buffer
			resp.Write(&buff)

			r.Reply(true, nil)
			tc.Channel.SendRequest("response", false, buff.Bytes())

			slog.Info("Response written back to proxy", "status", resp.StatusCode)
		}
	}
}

func (tc *TunnelClient) Listen() {
	tc.openRegistryConnection()
	defer tc.Close()

	tc.ForwardRequests()
}
