package sshhandler

type SSHTailConfig struct {
	Username    string             `json:"username"`    // Username used for executing the ssh tunnel
	Port        int32              `json:"port"`        // Port whitelisted for ssh tunnel
	Host        string             `json:"host"`        // Machine IP to ssh and fetch the logs
	Alias       bool               `json:"aliased"`     // If there is an alias exists in the config
	AliasString string             `json:"aliasString"` // Alias string that need to be used for ssh tunnel
	Commands    []ExecutionCommand `json:"commands"`    // Array of files to tail
}

type ExecutionCommand struct {
	CommandStr string `json:"command"` // Command that need to be executed on the terminal
	Outfile    string `json:"file"`    // Explicit file on file system to dump the Stdio from ssh session
}
