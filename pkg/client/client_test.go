package client

import (
	"log/slog"
	"testing"
)

func TestClientConfig(t *testing.T) {
	logger := slog.Default()

	if logger == nil {
		t.Fatal("Logger is nil")
	}
}
