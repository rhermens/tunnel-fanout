package registry

import (
	"errors"
	"log/slog"
	"net"
	"slices"
	"sync"

	"golang.org/x/crypto/ssh"
)

type OpenChannel struct {
	Type     ConnectionType
	Channel  ssh.Channel
	Requests <-chan *ssh.Request
}

type ConnectionType string

const (
	None   ConnectionType = "none"
	Client ConnectionType = "client"
	Proxy  ConnectionType = "proxy"
)

type ProxyRequestType string

const (
	Forward ProxyRequestType = "forward"
)

type Connection struct {
	Type            ConnectionType
	SSHConn         *ssh.ServerConn
	ChannelRequests <-chan ssh.NewChannel
	Reqs            <-chan *ssh.Request
	OpenChannels    []*OpenChannel
	wg              sync.WaitGroup
}

func (r *Registry) AddConnection(nConn net.Conn) error {
	c, err := NewConnectionFromTCP(nConn, r.Config)
	if err != nil {
		return err
	}

	slog.Info("Accepted connection", "remote", c.SSHConn.RemoteAddr(), "local", c.SSHConn.LocalAddr())
	r.Connections[c.RemoteAddr().String()] = c

	c.wg.Go(func() {
		for openChan := range c.AcceptChannels() {
			if openChan.Type == Client {
				c.wg.Go(func() {
					ssh.DiscardRequests(openChan.Requests)
				})
			}

			if openChan.Type == Proxy {
				c.wg.Go(func() {
					r.HandleProxyChannelRequests(openChan.Requests)
				})
			}
		}
	})

	return nil
}

func (r *Registry) HandleProxyChannelRequests(reqs <-chan *ssh.Request) {
	for req := range reqs {
		if req.Type == string(Forward) {
			r.FanoutBuffer(req.Payload)
		}
		slog.Info("Received channel request", "type", req.Type, "want_reply", req.WantReply)
	}
}

func NewConnectionFromTCP(conn net.Conn, rConfig *RegistryServerConfig) (*Connection, error) {
	var err error
	c := &Connection{
		Type: None,
	}
	c.SSHConn, c.ChannelRequests, c.Reqs, err = ssh.NewServerConn(conn, rConfig.Ssh.SshConfig)
	if err != nil {
		slog.Error("Failed to handshake", "error", err)
		return nil, err
	}
	c.wg.Go(func() {
		ssh.DiscardRequests(c.Reqs)
	})

	return c, nil
}

func (c *Connection) Wait() error {
	return c.SSHConn.Wait()
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.SSHConn.RemoteAddr()
}

func (c *Connection) AcceptChannels() <-chan *OpenChannel {
	wg := sync.WaitGroup{}
	out := make(chan *OpenChannel)

	wg.Go(func() {
		for channel := range c.ChannelRequests {
			if channel.ChannelType() != string(Client) && channel.ChannelType() != string(Proxy) {
				channel.Reject(ssh.UnknownChannelType, "unsupported channel type")
				slog.Warn("Rejected channel request", "type", channel.ChannelType(), "remote", c.SSHConn.RemoteAddr(), "local", c.SSHConn.LocalAddr())
				continue
			}

			var err error
			openChannel := &OpenChannel{}
			openChannel.Channel, openChannel.Requests, err = channel.Accept()
			if err != nil {
				slog.Error("Could not accept channel", "error", err)
				continue
			}

			if channel.ChannelType() == string(Client) {
				c.Type = Client
				openChannel.Type = Client
			}

			if channel.ChannelType() == string(Proxy) {
				c.Type = Proxy
				openChannel.Type = Proxy
			}

			slog.Info("Accepted channel req", "remote", c.SSHConn.RemoteAddr(), "local", c.SSHConn.LocalAddr(), "type", c.Type)
			c.OpenChannels = append(c.OpenChannels, openChannel)

			out <- openChannel
		}
	})

	return out
}

func (c *Connection) Close() {
	for _, openChan := range c.OpenChannels {
		openChan.Channel.Close()
	}
	c.SSHConn.Close()
}

func (c *Connection) ForwardBuffer(buf []byte) error {
	slog.Info("Forwarding request to client", "remote", c.SSHConn.RemoteAddr(), "local", c.SSHConn.LocalAddr(), "channels", len(c.OpenChannels))
	for i, openConn := range c.OpenChannels {
		slog.Info("Writing to channel", "channel", i)
		reply, err := openConn.Channel.SendRequest("consume", true, buf)
		slog.Info("Request sent", "reply", reply)

		if err != nil {
			openConn.Channel.Close()
			c.OpenChannels = slices.Delete(c.OpenChannels, i, 1)
			slog.Error("Failed to send request, closing channel", "error", err)
		}
	}

	if len(c.OpenChannels) == 0 {
		return errors.New("no open channels")
	}

	return nil
}
