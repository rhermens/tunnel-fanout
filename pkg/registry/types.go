package registry

import (
	"golang.org/x/crypto/ssh"
)

type OpenChannel struct {
	Channel  ssh.Channel
	Requests <-chan *ssh.Request
}

type TunnelClient struct {
	SSHConn         *ssh.ServerConn
	ChannelRequests <-chan ssh.NewChannel
	Reqs            <-chan *ssh.Request
	OpenChannels    []OpenChannel
}
