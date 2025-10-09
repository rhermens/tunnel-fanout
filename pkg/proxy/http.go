package proxy

import (
	"bytes"
	"log/slog"
	"net"
	"net/http"

	"github.com/rhermens/tunnel-fanout/pkg/registry"
)

func Listen(config *HttpServerConfig) error {
	mux := http.NewServeMux()

	for _, handlePath := range config.Paths {
		mux.HandleFunc(handlePath, forwardHandler)
		slog.Info("Registered handler", "path", handlePath)
	}

	slog.Info("Starting http server", "host", config.Host, "port", config.Port)
	return http.ListenAndServe(net.JoinHostPort(config.Host, config.Port), mux)
}

func forwardHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received request", "method", r.Method, "url", r.URL.String())
	var buffer bytes.Buffer
	r.Write(&buffer)

	for _, tunnelClient := range registry.Upstreams {
		go func() {
			slog.Info("Forwarding request to upstream", "remote", tunnelClient.SSHConn.RemoteAddr(), "local", tunnelClient.SSHConn.LocalAddr(), "channels", len(tunnelClient.OpenChannels))
			for i, openConn := range tunnelClient.OpenChannels {
				slog.Info("Writing to channel", "channel", i)
				reply, err := openConn.Channel.SendRequest("forward", true, buffer.Bytes())
				if err != nil {
					slog.Error("Failed to send request", "error", err)
					registry.RemoveUpstream(tunnelClient, err)
				}

				slog.Info("Request sent", "reply", reply)
			}
		}()
	}

	w.Write([]byte("OK"))
}
