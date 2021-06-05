package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"
	"time"

	sshhandler "github.com/yashLadha/ssh-tail/sshHandling"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func determinePrivateKey(path string, passphrase string) (ssh.Signer, error) {
	key, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
		return nil, err
	}
	// Create the Signer for this private key.
	if passphrase == sshhandler.EMPTY_STRING {
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
	var prefix string
	if sshConfig.Unique {
		prefix = time.Now().Format(time.RFC3339)
	}
	for _, command := range sshConfig.Commands {
		go sshhandler.CommandExecution(client, sshhandler.ExecutionCommandArgs{
			Command: command,
			Prefix:  prefix,
		}, &wg)
	}
	wg.Wait()
}

func decideJSONConfig() string {
	return os.Getenv("SSH_TAIL_CONFIG")
}

func main() {
	jsonConfig := fetchJSONConfig()
	var sshConfig sshhandler.SSHTailConfig
	sshConfig = sshhandler.SSHConfigParsing(jsonConfig)
	var sshDir string
	sshDir = sshhandler.GetSSHDir()
	hostkeyCallback := determineHosts(sshDir)
	signer := fetchPrivateKey(sshDir, sshConfig)
	config := createSSHConfig(sshConfig, signer, hostkeyCallback)
	var machineIP string
	machineIP = fmt.Sprintf("%s:%d", sshConfig.Host, sshConfig.Port)
	log.Printf("Initiating connection to %s", machineIP)
	client := sshPublicConnection(machineIP, config)
	defer client.Close()

	processCommands(client, sshConfig)
}

func determineHosts(sshDir string) ssh.HostKeyCallback {
	hostkeyCallback, err := determineHostsCallback(path.Join(sshDir, "known_hosts"))
	if err != nil {
		log.Fatalf("Unable to parse the hosts file: %v", err)
	}
	return hostkeyCallback
}

func fetchJSONConfig() string {
	jsonConfig := decideJSONConfig()
	if jsonConfig == sshhandler.EMPTY_STRING {
		log.Fatalf("ENV variable SSH_TAIL_CONFIG is not set")
	}
	return jsonConfig
}

func sshPublicConnection(machineIP string, config *ssh.ClientConfig) *ssh.Client {
	client, err := ssh.Dial("tcp", machineIP, config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	log.Printf("Connetion setup to %s", machineIP)
	return client
}

func createSSHConfig(sshConfig sshhandler.SSHTailConfig, signer ssh.Signer, hostkeyCallback ssh.HostKeyCallback) *ssh.ClientConfig {
	config := &ssh.ClientConfig{
		User: sshConfig.Username,
		Auth: []ssh.AuthMethod{

			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostkeyCallback,
	}
	return config
}

func fetchPrivateKey(sshDir string, sshConfig sshhandler.SSHTailConfig) ssh.Signer {
	signer, err := determinePrivateKey(path.Join(sshDir, "id_rsa"), sshConfig.KeyPassPhrase)
	if err != nil {
		log.Fatalf("Unable to prepare private key: %v", err)
	}
	return signer
}
