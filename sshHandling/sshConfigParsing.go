package sshhandler

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
)

func GetSSHDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Unable to get home dir for user: %v", err)
	}
	return path.Join(dir, ".ssh")
}

func SSHConfigParsing(configPath string) SSHTailConfig {
	jsonFile, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Unable to open json config file %v", err)
	}

	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatalf("Unable to read json config file %v", err)
	}
	var sshTailConfig SSHTailConfig
	json.Unmarshal(byteValue, &sshTailConfig)

	return sshTailConfig
}
