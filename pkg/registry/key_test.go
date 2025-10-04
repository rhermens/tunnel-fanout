package registry

import (
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestKeyGen(t *testing.T) {
	pem, err := GenerateHostKey("/tmp/id_rsa_test")
	if err != nil {
		t.Fatal(err)
	}

	privateKey, err := ssh.ParsePrivateKey(pem)

	if err != nil {
		t.Fatal(err)
	}

	if privateKey == nil {
		t.Fatal("privateKey is nil")
	}
}
