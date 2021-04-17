## ssh-tail

Simple binary to tail log files from remote ssh machines and store them locally
into your system for debugging or storage purpose. Currently working on
implementing the local file system sink.

It assumes that you have an unecrypted private key and present in your home
folder inside an `.ssh` folder.

### Sinks
Sinks are the interfaces which will be used to dump the data fetched from the
ssh session running on the remote machine. These sinks can be local file system
or an external service like S3. For now the implementation is only made for the
local file system but later S3 or any plugin can be included.
