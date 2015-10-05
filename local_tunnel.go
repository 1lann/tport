package main

import (
	"io"
	"net"
	"strings"
)

var localTunnels map[string]*tunnel = make(map[string]*tunnel)

func openLocalTunnel(localPort, remotePort string) (*tunnel, error) {
	remoteListener, err := connectedClient.Listen("tcp", "0.0.0.0:"+remotePort)
	if err != nil {
		return nil, err
	}

	thisTunnel := &tunnel{
		open:        true,
		localPort:   localPort,
		remotePort:  remotePort,
		listener:    remoteListener,
		connections: make(map[string]connectionPair),
	}

	go func(remoteListener net.Listener, thisTunnel *tunnel) {
		for {
			conn, err := remoteListener.Accept()
			if err != nil {
				if strings.Contains(err.Error(),
					"use of closed network connection") || err == io.EOF {
					return
				}

				continue
			}

			go handleLocalTunnelConn(thisTunnel, conn)
		}
	}(remoteListener, thisTunnel)

	return thisTunnel, nil
}

func handleLocalTunnelConn(thisTunnel *tunnel, remoteConn net.Conn) {
	connChannel := make(chan bool)

	for thisTunnel.open {
		localConn, err := net.Dial("tcp", "127.0.0.1:"+thisTunnel.localPort)
		if err != nil {
			if strings.Contains(err.Error(),
				"can't assign requested address") {
				connectedClient.Close()
				return
			}

			remoteConn.Close()
			return
		}

		connPair := connectionPair{
			local:  localConn,
			remote: remoteConn,
		}

		thisTunnel.connections[remoteConn.RemoteAddr().String()] = connPair

		go func() {
			_, err := io.Copy(remoteConn, localConn)
			if err != nil {
				if strings.Contains(err.Error(),
					"use of closed network connection") {
					connChannel <- false
					return
				}
			}

			connChannel <- true
		}()

		go func() {
			_, err := io.Copy(localConn, remoteConn)
			if err != nil {
				if strings.Contains(err.Error(),
					"use of closed network connection") {
					connChannel <- false
					return
				}
			}

			connChannel <- true
		}()

		shouldContinue := <-connChannel

		remoteConn.Close()
		localConn.Close()

		if !shouldContinue {
			break
		}
	}
}
