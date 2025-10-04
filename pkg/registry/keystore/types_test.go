package keystore

import (
	"testing"

	"golang.org/x/crypto/ssh"
)

var key1 = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEz6uCbKyV+dBXEM9nGmeEIpcvWJIitFe/Gq7yH5ucR0"
var key2 = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBKGAnuvRdrGRqFRkGuw8/dV0uH/Jd286NOceXgI45S8"
var key1P, _, _, _, _ = ssh.ParseAuthorizedKey([]byte(key1))
var key2P, _, _, _, _ = ssh.ParseAuthorizedKey([]byte(key2))

func TestParseAuthorizedKeysCommaSeperated(t *testing.T) {

	type testCase struct {
		name            string
		keys            []string
		expected        map[string]bool
		expectedMissing map[string]bool
	}

	testCases := []testCase{
		{
			name: "both keys",
			keys: []string{key1, key2},
			expected: map[string]bool{
				string(key1P.Marshal()): true,
				string(key2P.Marshal()): true,
			},
			expectedMissing: map[string]bool{},
		},
		{
			name: "single key",
			keys: []string{key1},
			expected: map[string]bool{
				string(key1P.Marshal()): true,
			},
			expectedMissing: map[string]bool{
				string(key2P.Marshal()): true,
			},
		},
		{
			name:     "empty keys",
			keys:     []string{},
			expected: map[string]bool{},
			expectedMissing: map[string]bool{
				string(key1P.Marshal()): true,
				string(key2P.Marshal()): true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewFromStrings(tc.keys)
			if len(actual) != len(tc.expected) {
				t.Fatalf("Expected %d authorized keys, got %d", len(tc.expected), len(actual))
			}

			for key := range tc.expected {
				if !actual[key] {
					t.Errorf("Expected key %s to be present", key)
				}
			}
			for key := range tc.expectedMissing {
				if actual[key] {
					t.Errorf("Expected key %s not to be present", key)
				}
			}
		})
	}
}
