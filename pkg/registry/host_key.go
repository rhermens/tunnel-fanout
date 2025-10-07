package registry

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log/slog"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

func GenerateHostKey(p string) ([]byte, error) {
	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	privateBytes, err := x509.MarshalPKCS8PrivateKey(rsaPrivateKey)
	if err != nil {
		return nil, err
	}

	privateKeyPemBytes := pem.EncodeToMemory(&pem.Block{
		Type:    "PRIVATE KEY",
		Headers: nil,
		Bytes:   privateBytes,
	})

	return privateKeyPemBytes, nil
}

func WriteHostKey(p string, privateKeyPemBytes []byte) error {
	slog.Info("Writing new host key", "path", p)
	err := os.MkdirAll(filepath.Dir(p), 0700)
	if err != nil {
		return err
	}

	return os.WriteFile(p, privateKeyPemBytes, 0600)
}

func EnsureHostKey(p string) ssh.Signer {
	var err error
	var pkBytes []byte
	if pkBytes, err = os.ReadFile(p); os.IsNotExist(err) {
		pkBytes, err = GenerateHostKey(p)
		err = WriteHostKey(p, pkBytes)

		if err != nil {
			slog.Error("Failed to generate host key", "error", err)
			panic(err)
		}
	}

	pk, err := ssh.ParsePrivateKey(pkBytes)
	if err != nil {
		slog.Error("Failed to parse host key", "error", err)
		panic(err)
	}

	return pk
}
