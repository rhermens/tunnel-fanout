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

	tc.Channel, tc.Requests, err = tc.RegistryClient.OpenChannel("direct-tcpip", []byte{})
	if err != nil {
		panic(err)
	}
	slog.Info("Opened channel to registry")
}

func (tc *TunnelClient) Close() {
	tc.RegistryClient.Close()
}

func (tc *TunnelClient) Listen() {
	tc.openRegistryConnection()
	defer tc.Close()

	cl := http.Client{}

	for r := range tc.Requests {
		slog.Info("Request", "type", r.Type, "want-reply", r.WantReply)
		reader := bytes.NewReader(r.Payload)
		hReq, err := http.ReadRequest(bufio.NewReader(reader))
		if err != nil {
			slog.Error("Failed to read request", "error", err)
			continue
		}

		u, err := url.Parse(fmt.Sprintf("http://localhost:%d%s", tc.LocalTargetPort, hReq.URL.Path))

		hReq.URL = u
		hReq.RequestURI = ""

		slog.Info("Forwarding request", "method", hReq.Method, "url", hReq.URL.String(), "host", hReq.Host)
		resp, err := cl.Do(hReq)
		if err != nil {
			slog.Error("Failed to do request", "error", err)
			continue
		}

		slog.Info("Forwarded req", "status", resp.StatusCode)
	}
}
