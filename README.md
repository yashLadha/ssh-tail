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
into your system for debugging or storage purpose. It is controlled by a JSON
config which saves me time of creating multiple ssh-session either using a
window managers or some bash script which does that for me. Logic for creating
new filenames based on the unique flag.

It assumes that you have an unecrypted private key and present in your home
folder inside an `.ssh` folder and `ssh-agent` is running on your machine.

To use ssh-tail you need to set and env variable while running the process
`SSH_TAIL_CONFIG` which will pick the json config to use for the ssh session.

### Sinks

Sinks are the interfaces which will be used to dump the data fetched from the
ssh session running on the remote machine. These sinks can be local file system
or an external service like S3. For now the implementation is only made for the
local file system but later S3 or any plugin can be included.

### Usage

```sh
SSH_TAIL_CONFIG="ssh_tunnel.json" make
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

For use with machines that are behind a proxy:

```json
{
  "host": "target_machine_ip",
  "port": target_machine_port,
  "username": "username",
  "proxyConfig": {
    "host": "proxy_machine_ip / hop_ip",
    "port": proxy_jump_machine_port,
    "username": "proxy_jump_username"
  },
  "commands": [
    {
      "command": "command_1",
      "file": file_1
    }
    ...
  ]
}
```

## Building locally

You can build the binary locally using the following command.

```sh
make build
```

This will generate build files for targets (windows, linux and darwin).

## Development / Contribution

This is just a binary that i created for myself to ease my workflow. In case you
also face similar issues or want to solve some existing issue, feel free to
dive right in and send a patch my way. I would be happy to review your patch.

To test out the changes locally you can use the following command.

```sh
SSH_TAIL_CONFIG="ssh_tunnel.json" make
```