package sshhandler

import (
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/crypto/ssh"
)

func killSignalHandler(session *ssh.Session, wg *sync.WaitGroup) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGKILL)
	// Waiting for the sigkill signal to stop the waiting group
	<-sigc
	log.Printf("Stop handler called for the process. Stopping ssh session")

	err := session.Close()
	if err != nil {
		log.Printf("Error while closing the ssh session %v", err)
	}
	wg.Done()
}

func CommandExecution(client *ssh.Client, commandArgs ExecutionCommandArgs, wg *sync.WaitGroup) {
	command := commandArgs.Command
	prefix := commandArgs.Prefix
	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	go killSignalHandler(session, wg)

	var sink io.Writer
	if command.Outfile != EMPTY_STRING {
		fileName := command.Outfile
		if prefix != EMPTY_STRING {
			fileName = prefix + "-" + fileName
		}
		sink = LocalSink(fileName)
	} else {
		sink = os.Stdout
	}
	// Piping the response from the ssh session to the file and need an object which
	// implements io.Writer interface so that it can be used by the ssh session to
	// dump the output
	session.Stdout = sink
	if err := session.Run(command.CommandStr); err != nil {
		log.Fatalf("Failed to run: %v", err)
	}
}
