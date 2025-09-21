package listener

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
	LocalTargetPort int
	RegistryClient  *ssh.Client
	Channel         ssh.Channel
	Requests        <-chan *ssh.Request
}

func (tc *TunnelClient) openRegistryConnection() {
	var err error
	clientConfig := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		User:            "Ayyy",
		Auth: []ssh.AuthMethod{
			ssh.Password("Ayy"),
		},
	}

	tc.RegistryClient, err = ssh.Dial("tcp", ":8000", clientConfig)
	if err != nil {
		panic(err)
	}

	tc.Channel, tc.Requests, err = tc.RegistryClient.OpenChannel("upstream", []byte{})
	if err != nil {
		panic(err)
	}
	slog.Info("Opened channel to registry")
}

func (tc *TunnelClient) Close() {
	tc.RegistryClient.Close()
}

func (tc *TunnelClient) ForwardRequests() {
	cl := http.Client{}
	defer tc.Close()

	for r := range tc.Requests {
		reader := bytes.NewReader(r.Payload)
		hReq, err := http.ReadRequest(bufio.NewReader(reader))
		if err != nil {
			slog.Error("Failed to read request", "error", err)
			continue
		}

		u, err := url.Parse(fmt.Sprintf("http://localhost:%d%s", tc.LocalTargetPort, hReq.URL.Path))
		hReq.URL = u
		hReq.RequestURI = ""

		resp, err := cl.Do(hReq)
		if err != nil {
			slog.Error("Failed to do request", "error", err)
			continue
		}

		slog.Info("Request forwarded", "path", hReq.URL.Path, "status", resp.StatusCode)
	}
}

func (tc *TunnelClient) Listen() {
	tc.openRegistryConnection()
	defer tc.Close()

	tc.ForwardRequests()
}
