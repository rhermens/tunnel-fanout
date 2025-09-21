package proxy

import (
	"bytes"
	"log/slog"
	"net"
	"net/http"

	"github.com/rhermens/tunnel-fanout/pkg/registry"
)

func Listen(config *HttpServerConfig) {
	mux := http.NewServeMux()

	mux.HandleFunc("/{path...}", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Received request", "method", r.Method, "url", r.URL.String())
		var buffer bytes.Buffer
		r.Write(&buffer)

		for _, tunnelClient := range registry.Upstreams {
			slog.Info("Forwarding request to upstream", "remote", tunnelClient.SSHConn.RemoteAddr(), "local", tunnelClient.SSHConn.LocalAddr(), "channels", len(tunnelClient.OpenChannels))

			for i, openConn := range tunnelClient.OpenChannels {
				slog.Info("Writing to channel", "channel", i)
				openConn.Channel.SendRequest("forward", false, buffer.Bytes())
			}
		}

		w.Write([]byte("OK"))
	})

	slog.Info("Starting http server", "host", config.Host, "port", config.Port)
	err := http.ListenAndServe(net.JoinHostPort(config.Host, config.Port), mux)

	if err != nil {
		slog.Error("Server failed", "error", err)
	}
}
