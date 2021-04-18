## ssh-tail

<p align="center">
  <img src="./assets/logo.png">
</p>

This project is one of the problems that I generally face while debugging some
system. When I am reproducing the issue on the machine i also want to tail the
logs side-by-side to see the error. This becomes cumbersome when i need to fetch
multiple log files or use them for future purpose or constantly hop between
dashboards, script and machine for the session.

To solve this comes `ssh-tail`. It is a
binary to tail log files from remote ssh machines and store them locally
into your system for debugging or storage purpose.

It assumes that you have an unecrypted private key and present in your home
folder inside an `.ssh` folder.

To use ssh-tail you need to set and env variable while running the process
`SSH_TAIL_CONFIG` which will pick the json config to use for the ssh session.

### Sinks
Sinks are the interfaces which will be used to dump the data fetched from the
ssh session running on the remote machine. These sinks can be local file system
or an external service like S3. For now the implementation is only made for the
local file system but later S3 or any plugin can be included.

### Usage
```sh
❯ SSH_TAIL_CONFIG="ssh_tunnel.json" make
```

Example cofig file:

```json
{
  "host": "machine_ip",
  "port": int_port_number,
  "username": "username",
  "commands": [
    {
      "command": "command_1",
      "file": "file_1"
    },
    {
      "command": "command_2",
      "file": "file_2"
    }
    ...
  ]
}
```

## Building locally
You can build the binary locally using the following command.

```sh
❯ make build
```

This will generate build files for targets (windows, linux and darwin).

## Development / Contribution

This is just a binary that i created for myself to ease my workflow. In case you
also face similar issues or want to solve some exisisting issue, feel free to
dive right in and send a patch my way. I would be happy to review your patch.

To test out the changes locally you can use the following command.

```sh
❯ SSH_TAIL_CONFIG="ssh_tunnel.json" make
```

## TODO
* Allow support for encrypted private keys
* Allow support for keys with passphrase
