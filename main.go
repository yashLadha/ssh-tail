package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"

	sshhandler "github.com/yashLadha/ssh-tail/sshHandling"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const EMPTY_STRING string = ""

func determinePrivateKey(path string) (ssh.Signer, error) {
	// A public key may be used to authenticate against the remote
	// server by using an unencrypted PEM-encoded private key file.
	//
	// If you have an encrypted private key, the crypto/x509 package
	// can be used to decrypt it.
	key, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
		return nil, err
	}
	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
		return nil, err
	}
	return signer, err
}

func determineHostsCallback(path string) (ssh.HostKeyCallback, error) {
	hostkeyCallback, err := knownhosts.New(path)
	if err != nil {
		log.Fatalf("Unable to read hosts file: %v", err)
	}
	return hostkeyCallback, err
}

func processCommands(client *ssh.Client, sshConfig sshhandler.SSHTailConfig) {
	var wg sync.WaitGroup
	wg.Add(len(sshConfig.Commands))
	for _, command := range sshConfig.Commands {
		go sshhandler.CommandExecution(client, command, &wg)
	}
	wg.Wait()
}

func decideJSONConfig() string {
	return os.Getenv("SSH_TAIL_CONFIG")
}

func main() {
	jsonConfig := decideJSONConfig()
	if jsonConfig == EMPTY_STRING {
		log.Fatalf("ENV variable SSH_TAIL_CONFIG is not set")
	}
	var sshConfig sshhandler.SSHTailConfig
	sshConfig = sshhandler.SSHConfigParsing(jsonConfig)
	var sshDir string
	sshDir = sshhandler.GetSSHDir()
	hostkeyCallback, err := determineHostsCallback(path.Join(sshDir, "known_hosts"))
	if err != nil {
		log.Fatalf("Unable to parse the hosts file: %v", err)
	}
	signer, err := determinePrivateKey(path.Join(sshDir, "id_rsa"))
	if err != nil {
		log.Fatalf("Unable to prepare private key: %v", err)
	}
	config := &ssh.ClientConfig{
		User: sshConfig.Username,
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostkeyCallback,
	}
	var machineIP string
	machineIP = fmt.Sprintf("%s:%d", sshConfig.Host, sshConfig.Port)
	log.Printf("Initiating connection to %s", machineIP)
	client, err := ssh.Dial("tcp", machineIP, config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	log.Printf("Connetion setup to %s", machineIP)
	defer client.Close()

	processCommands(client, sshConfig)
}
