package main

import (
	"io"
	"net"
	"strings"
)

var remoteTunnels map[string]*tunnel = make(map[string]*tunnel)

func openRemoteTunnel(localPort, remotePort string) (*tunnel, error) {
	localListener, err := net.Listen("tcp", "127.0.0.1:"+localPort)
	if err != nil {
		return nil, err
	}

	thisTunnel := &tunnel{
		open:        true,
		localPort:   localPort,
		remotePort:  remotePort,
		listener:    localListener,
		connections: make(map[string]connectionPair),
	}

	go func(localListener net.Listener, thisTunnel *tunnel) {
		for {
			conn, err := localListener.Accept()
			if err != nil {
				if strings.Contains(err.Error(),
					"use of closed network connection") || err == io.EOF {
					return
				}

				continue
			}

			go handleRemoteTunnelConn(thisTunnel, conn)
		}
	}(localListener, thisTunnel)

	return thisTunnel, nil
}

func handleRemoteTunnelConn(thisTunnel *tunnel, localConn net.Conn) {
	connChannel := make(chan bool)

	for thisTunnel.open {
		remoteConn, err := connectedClient.Dial("tcp", "127.0.0.1:"+
			thisTunnel.remotePort)
		if err != nil {
			if strings.Contains(err.Error(),
				"can't assign requested address") {
				connectedClient.Close()
				return
			}

			localConn.Close()
			return
		}

		connPair := connectionPair{
			local:  localConn,
			remote: remoteConn,
		}

		thisTunnel.connections[localConn.RemoteAddr().String()] = connPair

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

		shouldContinue := <-connChannel

		remoteConn.Close()
		localConn.Close()

		if !shouldContinue {
			break
		}
	}
}
