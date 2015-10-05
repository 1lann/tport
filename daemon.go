package main

import (
	"encoding/hex"
	"fmt"
	"github.com/VividCortex/godaemon"
	"golang.org/x/crypto/ssh"
	"net"
	"os"
	"os/signal"
	"strings"
)

const daemonPort = "15453"

var connectedUsername string
var connectedPassword string
var connectedHost string
var connectedClient *ssh.Client
var daemonListener net.Listener

type connectionPair struct {
	local  net.Conn
	remote net.Conn
}

type tunnel struct {
	open        bool
	localPort   string
	remotePort  string
	listener    net.Listener
	connections map[string]connectionPair
}

func (tun *tunnel) close() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Panic recovered in tunnel close: ", r)
		}
	}()

	tun.open = false
	for _, connection := range tun.connections {
		connection.local.Close()
		connection.remote.Close()
	}

	for key, thisTunnel := range remoteTunnels {
		if thisTunnel == tun {
			delete(remoteTunnels, key)
			break
		}
	}

	for key, thisTunnel := range localTunnels {
		if thisTunnel == tun {
			delete(localTunnels, key)
			break
		}
	}

	tun.listener.Close()
}

func newDaemon(username string, password string,
	host string, client *ssh.Client) {
	connectedUsername = username
	connectedPassword = password
	connectedHost = host
	connectedClient = client

	go sshConnectionMonitor()

	var err error
	daemonListener, err = net.Listen("tcp", ":"+daemonPort)
	if err != nil {
		fmt.Println("Sorry, the daemon could not bind to communications port " +
			daemonPort + ".")
		fmt.Println("Make sure there are no other applications which are using port " +
			daemonPort + ".")
		fmt.Println(err)
		return
	}

	fmt.Println("Connected! You may now open tunneled ports.")

	if len(os.Args) < 3 {
		os.Args = append(os.Args, hex.EncodeToString([]byte(password)))
		godaemon.MakeDaemon(&godaemon.DaemonAttr{})
	}

	go signalHandler()

	for {
		conn, err := daemonListener.Accept()
		if err != nil {
			if strings.Contains(err.Error(),
				"use of closed network connection") {
				return
			}

			fmt.Println("Error accepting connection ", err)
			continue
		}

		go handleRequest(conn)
	}
}

func closeDaemon() {
	for _, tun := range localTunnels {
		tun.close()
	}

	for _, tun := range remoteTunnels {
		tun.close()
	}

	connectedClient.Close()
	daemonListener.Close()
}

func signalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Println("\nShutting down daemon...")
	closeDaemon()
}

func sshConnectionMonitor() {
	for {
		err := connectedClient.Wait()

		if err != nil && strings.Contains(err.Error(),
			"use of closed network connection") {
			return
		}

		fmt.Println("Remote connection unexpectedly closed.")
		fmt.Println(err)

		connectedClient, err = connectToRemote(connectedUsername,
			connectedPassword, connectedHost)
		if err != nil {
			fmt.Println("Could not reconnect to remote server!")
			fmt.Println("Shutting down daemon...")
			closeDaemon()
			return
		}

		return
	}
}

func handleRequest(conn net.Conn) {
	for {
		resp, err := readResponse(conn)
		if err != nil {
			fmt.Println("Read error: ", err)
			conn.Close()
			return
		}

		request := strings.Split(resp, " ")

		if request[0] == "tport" {
			switch request[1] {
			case "hello":
				conn.Write([]byte("tport " + connectedUsername + "@" +
					connectedHost + "\r"))
			case "open":
				openConnectionRequest(conn, request)
				return
			case "close":
				closeConnectionRequest(conn, request)
				return
			case "list":
				listRequest(conn)
				return
			case "dc":
				conn.Write([]byte("Disconnected.\r"))
				closeDaemon()
			default:
				conn.Write([]byte("daemon: unknown command.\r"))
			}
		} else {
			fmt.Println("Unexpected message: " + resp)
		}
	}
}
