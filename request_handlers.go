package main

import (
	"net"
)

func openConnectionRequest(conn net.Conn, request []string) {
	defer conn.Close()

	if (request[2] != "remote" && request[2] != "local") ||
		(len(request) < 4) {
		conn.Write([]byte(
			"Invalid direction option. Must be remote or local.\r"))
		return
	}

	if !isNumber(request[3]) {
		conn.Write([]byte("Primary port must be a number.\r"))
		return
	}

	secondaryPort := request[3]

	if len(request) > 4 {
		if !isNumber(request[4]) {
			conn.Write([]byte("Secondary port must be a number.\r"))
			return
		}

		secondaryPort = request[4]
	}

	if request[2] == "remote" {
		// Remote request
		if _, found := remoteTunnels[secondaryPort]; found {
			conn.Write([]byte("A tunnel is already open at local port " +
				secondaryPort + ".\r"))
			return
		}

		tun, err := openRemoteTunnel(secondaryPort, request[3])
		if err != nil {
			conn.Write([]byte("Failed to open tunnel:\n" +
				"    " + err.Error() + "\r"))
		} else {
			remoteTunnels[secondaryPort] = tun
			conn.Write([]byte("Tunnel opened on local port " + secondaryPort + ".\r"))
		}
	} else {
		// Local request
		if _, found := localTunnels[secondaryPort]; found {
			conn.Write([]byte("A tunnel is already open at remote port " +
				secondaryPort + ".\r"))
			return
		}

		tun, err := openLocalTunnel(request[3], secondaryPort)
		if err != nil {
			conn.Write([]byte("Failed to open tunnel:\n" +
				"    " + err.Error() + "\r"))
		} else {
			localTunnels[secondaryPort] = tun
			conn.Write([]byte("Tunnel opened on remote port " + secondaryPort + ".\r"))
		}
	}
}

func closeConnectionRequest(conn net.Conn, request []string) {
	defer conn.Close()

	if request[2] == "all" {
		for _, tun := range localTunnels {
			tun.close()
		}

		for _, tun := range remoteTunnels {
			tun.close()
		}

		conn.Write([]byte("All tunnels closed.\r"))
	} else if request[2] == "remote" {
		if len(request) < 4 {
			conn.Write([]byte("The local listening port must be specified.\r"))
			return
		}

		if request[3] == "all" {
			for _, tun := range remoteTunnels {
				tun.close()
			}

			conn.Write([]byte("All remote (local -> remote) tunnels closed.\r"))
			return
		}

		if tun, found := remoteTunnels[request[3]]; found {
			tun.close()
			conn.Write([]byte("Tunnel with local listening port " +
				request[3] + " closed.\r"))
		} else {
			conn.Write([]byte("No open tunnels with local listening port " +
				request[3] + " found.\r"))
			return
		}
	} else if request[2] == "local" {
		if len(request) < 4 {
			conn.Write([]byte("The remote listening port must be specified.\r"))
			return
		}

		if request[3] == "all" {
			for _, tun := range localTunnels {
				tun.close()
			}

			conn.Write([]byte("All local (remote -> local) tunnels closed.\r"))
			return
		}

		if tun, found := localTunnels[request[3]]; found {
			tun.close()
			conn.Write([]byte("Tunnel with remote listening port " +
				request[3] + " closed.\r"))
		} else {
			conn.Write([]byte("No open tunnels with remote listening port " +
				request[3] + " found.\r"))
			return
		}
	} else {
		conn.Write([]byte(
			"Specify whether the tunnel is remote, local or all.\r"))
	}
}

func listRequest(conn net.Conn) {
	if len(remoteTunnels) > 0 {
		conn.Write([]byte("Remote tunnels : local -> remote\n"))
		for _, tun := range remoteTunnels {
			conn.Write([]byte("    " + tun.localPort + " -> " +
				tun.remotePort + "\n"))
		}
	} else {
		conn.Write([]byte("No remote tunnels open.\n"))
	}

	if len(localTunnels) > 0 {
		conn.Write([]byte("Local tunnels : remote -> local\n"))
		for _, tun := range localTunnels {
			conn.Write([]byte("    " + tun.remotePort + " -> " +
				tun.localPort + "\n"))
		}
	} else {
		conn.Write([]byte("No local tunnels open."))
	}

	conn.Write([]byte{'\r'})
}
