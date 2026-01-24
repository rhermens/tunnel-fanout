package registry

import (
	"net"
	"strings"
	"testing"

	"github.com/rhermens/tunneld/pkg/registry/keystore"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

var key1 = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEz6uCbKyV+dBXEM9nGmeEIpcvWJIitFe/Gq7yH5ucR0"
var key2 = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBKGAnuvRdrGRqFRkGuw8/dV0uH/Jd286NOceXgI45S8"
var key1P, _, _, _, _ = ssh.ParseAuthorizedKey([]byte(key1))
var key2P, _, _, _, _ = ssh.ParseAuthorizedKey([]byte(key2))

func setup() {
	viper.SetEnvPrefix("tunneld")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

type mockSshConnMeta struct{}

func (m *mockSshConnMeta) User() string          { return "testuser" }
func (m *mockSshConnMeta) RemoteAddr() net.Addr  { return &net.IPAddr{} }
func (m *mockSshConnMeta) LocalAddr() net.Addr   { return &net.IPAddr{} }
func (m *mockSshConnMeta) SessionID() []byte     { return []byte("sessionid") }
func (m *mockSshConnMeta) ClientVersion() []byte { return []byte("clientversion") }
func (m *mockSshConnMeta) ServerVersion() []byte { return []byte("serverversion") }

func TestPublicKeyCallback(t *testing.T) {
	setup()

	type testCase struct {
		name                string
		authorizedKeys      keystore.Keystore
		keyToTest           ssh.PublicKey
		expectedPermissions bool
		expectedError       bool
	}

	testCases := []testCase{
		{
			name: "key is authorized",
			authorizedKeys: keystore.Keystore{
				string(key1P.Marshal()): true,
			},
			keyToTest:           key1P,
			expectedPermissions: true,
			expectedError:       false,
		},
		{
			name: "key is missing",
			authorizedKeys: keystore.Keystore{
				string(key2P.Marshal()): true,
			},
			keyToTest:           key1P,
			expectedPermissions: false,
			expectedError:       true,
		},
		{
			name:                "authorized keys is empty",
			authorizedKeys:      keystore.Keystore{},
			keyToTest:           key1P,
			expectedPermissions: false,
			expectedError:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			callback := newPublicKeyCallback(tc.authorizedKeys)
			permissions, err := callback(&mockSshConnMeta{}, tc.keyToTest)

			if err == nil && tc.expectedError {
				t.Fatalf("Expected key to be rejected, got permissions: %v", permissions)
			}

			if permissions == nil && tc.expectedPermissions {
				t.Fatalf("Expected key to be permitted, got error: %v", err)
			}
		})
	}

}
