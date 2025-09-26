package registry

import (
	"log/slog"
	"net"
	"sync"

	"golang.org/x/crypto/ssh"
)

type OpenChannel struct {
	Channel  ssh.Channel
	Requests <-chan *ssh.Request
}

type TunnelUpstream struct {
	SSHConn         *ssh.ServerConn
	ChannelRequests <-chan ssh.NewChannel
	Reqs            <-chan *ssh.Request
	OpenChannels    []*OpenChannel
	wg              sync.WaitGroup
}

func NewUpstreamFromTCP(conn net.Conn, rConfig *RegistryConfig) (*TunnelUpstream, error) {
	var err error
	tu := &TunnelUpstream{}
	tu.SSHConn, tu.ChannelRequests, tu.Reqs, err = ssh.NewServerConn(conn, rConfig.Ssh.SshConfig)
	if err != nil {
		slog.Error("Failed to handshake", "error", err)
		return nil, err
	}

	tu.wg.Go(func() {
		ssh.DiscardRequests(tu.Reqs)
	})
	tu.wg.Go(func() {
		tu.acceptChannels()
	})

	return tu, nil
}

func (tu *TunnelUpstream) acceptChannels() {
	for channel := range tu.ChannelRequests {
		var err error
		openChannel := &OpenChannel{}
		openChannel.Channel, openChannel.Requests, err = channel.Accept()
		if err != nil {
			slog.Error("Could not accept channel", "error", err)
			continue
		}

		tu.wg.Go(func() {
			for req := range openChannel.Requests {
				slog.Info("Received channel request", "type", req.Type, "want_reply", req.WantReply)
			}
		})

		slog.Info("Accepted channel req", "remote", tu.SSHConn.RemoteAddr(), "local", tu.SSHConn.LocalAddr())
		tu.OpenChannels = append(tu.OpenChannels, openChannel)
	}
}

func (tu *TunnelUpstream) Close() {
	for _, openChan := range tu.OpenChannels {
		openChan.Channel.Close()
	}
	tu.SSHConn.Close()
}
