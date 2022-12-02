package sshhandler

import (
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"path"
)

func determinePrivateKey(path string, passphrase string) (ssh.Signer, error) {
	key, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
		return nil, err
	}
	// Create the Signer for this private key.
	if passphrase == EmptyString {
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			log.Fatalf("unable to parse private key: %v", err)
		}
		return signer, err
	}
	signer, err := ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
	if err != nil {
		log.Fatalf("unable to parse private key with passphrase: %v", err)
	}
	return signer, err
}

func FetchPrivateKey(sshDir string, sshConfig SSHTailConfig) ssh.Signer {
	signer, err := determinePrivateKey(path.Join(sshDir, "id_rsa"), sshConfig.KeyPassPhrase)
	if err != nil {
		log.Fatalf("Unable to prepare private key: %v", err)
	}
	return signer
}
