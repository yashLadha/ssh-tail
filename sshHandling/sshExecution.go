package sshhandler

import (
	"io"
	"log"
	"sync"

	"golang.org/x/crypto/ssh"
)

func CommandExecution(client *ssh.Client, command ExecutionCommand, wg *sync.WaitGroup) {
	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer wg.Done()
	defer session.Close()

	var sink io.Writer
	sink = LocalSink(command.Outfile)
	// Piping the response from the ssh session to the file and need an object which
	// implements io.Writer interface so that it can be used by the ssh session to
	// dump the output
	session.Stdout = sink
	if err := session.Run(command.CommandStr); err != nil {
		log.Fatalf("Failed to run: %v", err)
	}
}
