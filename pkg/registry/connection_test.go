package registry

import (
	"errors"
	"io"
	"net"
	"testing"
)

type MockChannel struct {
	MockSendRequest func(name string, wantReply bool, payload []byte) (bool, error)
}

type MockSSHConn struct{}

func (mc MockChannel) Close() error                   { return nil }
func (mc MockChannel) CloseWrite() error              { return nil }
func (mc MockChannel) Stderr() io.ReadWriter          { return nil }
func (mc MockChannel) Read(data []byte) (int, error)  { return 0, nil }
func (mc MockChannel) Write(data []byte) (int, error) { return len(data), nil }
func (mc MockChannel) SendRequest(name string, wantReply bool, payload []byte) (bool, error) {
	return mc.MockSendRequest(name, wantReply, payload)
}

func (msc MockSSHConn) Close() error         { return nil }
func (msc MockSSHConn) LocalAddr() net.Addr  { return nil }
func (msc MockSSHConn) RemoteAddr() net.Addr { return nil }
func (msc MockSSHConn) Wait() error          { return nil }

func TestForwardBuffer(t *testing.T) {
	type testCase struct {
		name        string
		mocks       []func(name string, wantReply bool, payload []byte) (bool, error)
		expectedLen int
	}

	tc := []testCase{
		{
			name: "All channels succeed",
			mocks: []func(name string, wantReply bool, payload []byte) (bool, error){
				func(name string, wantReply bool, payload []byte) (bool, error) {
					return true, nil
				},
				func(name string, wantReply bool, payload []byte) (bool, error) {
					return true, nil
				},
			},
			expectedLen: 2,
		},
		{
			name: "One channel fails",
			mocks: []func(name string, wantReply bool, payload []byte) (bool, error){
				func(name string, wantReply bool, payload []byte) (bool, error) {
					return true, nil
				},
				func(name string, wantReply bool, payload []byte) (bool, error) {
					return false, errors.New("Some error")
				},
			},
			expectedLen: 1,
		},
		{
			name: "All channels fail",
			mocks: []func(name string, wantReply bool, payload []byte) (bool, error){
				func(name string, wantReply bool, payload []byte) (bool, error) {
					return false, errors.New("Some error")
				},
				func(name string, wantReply bool, payload []byte) (bool, error) {
					return false, errors.New("Some error")
				},
			},
			expectedLen: 0,
		},
	}

	for _, test := range tc {
		channels := make([]*OpenChannel, len(test.mocks))
		for i, mock := range test.mocks {
			channels[i] = &OpenChannel{
				Type: Client,
				Channel: MockChannel{
					MockSendRequest: mock,
				},
			}
		}

		connection := &Connection{
			SSHConn:      MockSSHConn{},
			OpenChannels: channels,
		}

		connection.ForwardBuffer([]byte("test data"))
		if len(connection.OpenChannels) != test.expectedLen {
			t.Error("Expected all channels to be removed after errors")
		}
	}
}
