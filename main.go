package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	sshhandler "github.com/yashLadha/ssh-tail/sshHandling"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
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
	sshConfig := sshhandler.SSHConfigParsing(jsonConfig)
	sshDir := sshhandler.GetSSHDir()
	hostkeyCallback := determineHosts(sshDir)
	signer := fetchPrivateKey(sshDir, sshConfig)
	var client *ssh.Client
	if sshConfig.ProxyConfig != nil {
		client = privateSSH(sshConfig, signer, hostkeyCallback)
	} else {
		config := createSSHConfig(sshConfig, signer, hostkeyCallback)
		machineIP := fmt.Sprintf("%s:%d", sshConfig.Host, sshConfig.Port)
		log.Printf("Initiating connection to %s", machineIP)
		client = sshPublicConnection(machineIP, config)
	}
	defer client.Close()
	processCommands(client, sshConfig)
}

func privateSSH(sshConfig sshhandler.SSHTailConfig, signer ssh.Signer, hostkeyCallback ssh.HostKeyCallback) *ssh.Client {
	proxyConfig := createSSHConfig(*sshConfig.ProxyConfig, signer, hostkeyCallback)
	proxyMachineIP := net.JoinHostPort(sshConfig.ProxyConfig.Host, strconv.Itoa(int(sshConfig.ProxyConfig.Port)))
	log.Printf("Initiating connection to proxy: %s\n", proxyMachineIP)
	proxyClient, err := ssh.Dial("tcp", proxyMachineIP, proxyConfig)
	if err != nil {
		log.Fatalf("Error in setting up proxy connection %v", err)
	}
	machineIP := net.JoinHostPort(sshConfig.Host, strconv.Itoa(int(sshConfig.Port)))
	clientConn, err := proxyClient.Dial("tcp", machineIP)
	if err != nil {
		log.Fatalf("Error in dialing connection from proxy: %v", err)
	}
	targetConfig := createSSHConfig(sshConfig, signer, hostkeyCallback)
	ncc, chans, reqs, err := ssh.NewClientConn(clientConn, machineIP, targetConfig)
	if err != nil {
		log.Fatalf("Error in creating new client connection %v", err)
	}
	log.Printf("Connected %s\n", machineIP)
	return ssh.NewClient(ncc, chans, reqs)
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
	auths := []ssh.AuthMethod{}
	auths = append(auths, ssh.PublicKeys(signer))
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
	}
	config := &ssh.ClientConfig{
		User:            sshConfig.Username,
		Auth:            auths,
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
