package server

import (
	"bytes"
	"log/slog"
	"net/http"

	"github.com/rhermens/tunnel-fanout/pkg/registry"
)

func Listen(addr string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/{path...}", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Received request", "method", r.Method, "url", r.URL.String())
		var buffer bytes.Buffer
		r.Write(&buffer)

		for _, tunnelClient := range registry.TunnelClients {
			slog.Info("Forwarding request to listener", "remote", tunnelClient.SSHConn.RemoteAddr(), "local", tunnelClient.SSHConn.LocalAddr(), "channels", len(tunnelClient.OpenChannels))

			for i, openConn := range tunnelClient.OpenChannels {
				slog.Info("Writing to channel", "channel", i)
				openConn.Channel.SendRequest("forward", false, buffer.Bytes())
			}
		}

		w.Write([]byte("OK"))
	})

	slog.Info("Starting server", "addr", addr)
	err := http.ListenAndServe(addr, mux)

	if err != nil {
		slog.Error("Server failed", "error", err)
	}
}
