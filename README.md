# tport
tport is short for tunnel port. A CLI utility for port forwarding over SSH. Sign in once to your server, forward as many ports in either direction as you need.

## Usage
tport v0.1 by Jason Chu (1lann)

First, connect with
```
tport user@host
```

For tunnels from the local host, to the remote host.
```
tport remote remoteport [localport]
```
For tunnels from the remote host, to the local host.
```
tport local localport [remoteport]
```
```
tport close remote localport
tport close local remoteport
tport close remote/local all
tport close all
```

List all the open tunnels.
```
tport list
```
Disconnect, close all tunnels, and quit the daemon.
```
tport dc/disconnect/logout
```

## License
Licensed under the MIT License. See /LICENSE.
