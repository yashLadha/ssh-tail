package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	sshHandler "github.com/yashLadha/ssh-tail/sshHandling"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

func determineHostsCallback(path string) (ssh.HostKeyCallback, error) {
	hostKeyCallback, err := knownhosts.New(path)
	if err != nil {
		log.Fatalf("Unable to read hosts file: %v", err)
	}
	return hostKeyCallback, err
}

func processCommands(client *ssh.Client, sshConfig sshHandler.SSHTailConfig) {
	var wg sync.WaitGroup
	wg.Add(len(sshConfig.Commands))
	var prefix string
	if sshConfig.Unique {
		prefix = time.Now().Format(time.RFC3339)
	}
	for _, command := range sshConfig.Commands {
		go sshHandler.CommandExecution(client, sshHandler.ExecutionCommandArgs{
			Command: command,
			Prefix:  prefix,
		}, &wg)
	}
	wg.Wait()
}

func getSSHTailConfig() string {
	return os.Getenv("SSH_TAIL_CONFIG")
}

func main() {
	jsonConfig := fetchJSONConfig()
	sshConfig := sshHandler.SSHConfigParsing(jsonConfig)
	sshDir := sshHandler.GetSSHDir()
	hostKeyCallback := determineHosts(sshDir)
	signer := sshHandler.FetchPrivateKey(sshDir, sshConfig)
	var client *ssh.Client
	if sshConfig.ProxyConfig != nil {
		client = privateSSH(sshConfig, signer, hostKeyCallback)
	} else {
		config := createSSHConfig(sshConfig, signer, hostKeyCallback)
		machineIP := fmt.Sprintf("%s:%d", sshConfig.Host, sshConfig.Port)
		log.Printf("Initiating connection to %s", machineIP)
		client = sshPublicConnection(machineIP, config)
	}
	defer closeClientConnection(client)
	processCommands(client, sshConfig)
}

func closeClientConnection(client *ssh.Client) {
	err := client.Close()
	if err != nil {
		log.Fatalf("Error in closing connection %v\n", err)
	}
}

func privateSSH(sshConfig sshHandler.SSHTailConfig, signer ssh.Signer, hostKeyCallback ssh.HostKeyCallback) *ssh.Client {
	proxyConfig := createSSHConfig(*sshConfig.ProxyConfig, signer, hostKeyCallback)
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
	targetConfig := createSSHConfig(sshConfig, signer, hostKeyCallback)
	ncc, channels, reqs, err := ssh.NewClientConn(clientConn, machineIP, targetConfig)
	if err != nil {
		log.Fatalf("Error in creating new client connection %v", err)
	}
	log.Printf("Connected %s\n", machineIP)
	return ssh.NewClient(ncc, channels, reqs)
}

func determineHosts(sshDir string) ssh.HostKeyCallback {
	hostKeyCallback, err := determineHostsCallback(path.Join(sshDir, "known_hosts"))
	if err != nil {
		log.Fatalf("Unable to parse the hosts file: %v", err)
	}
	return hostKeyCallback
}

func fetchJSONConfig() string {
	jsonConfig := getSSHTailConfig()
	if jsonConfig == sshHandler.EmptyString {
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

func createSSHConfig(sshConfig sshHandler.SSHTailConfig, signer ssh.Signer, hostKeyCallback ssh.HostKeyCallback) *ssh.ClientConfig {
	var auths []ssh.AuthMethod
	auths = append(auths, ssh.PublicKeys(signer))
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
	}
	return &ssh.ClientConfig{
		User:            sshConfig.Username,
		Auth:            auths,
		HostKeyCallback: hostKeyCallback,
	}
}
