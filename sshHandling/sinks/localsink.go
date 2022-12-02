package sinks

import (
	"io"
	"log"
	"os"
)

func LocalSink(outFile string) io.Writer {
	fd, err := os.Create(outFile)
	if err != nil {
		log.Fatalf("Unable to create file %s %v", outFile, err)
	}
	return fd
}
