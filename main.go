package main

import (
	"encoding/hex"
	"fmt"
	"github.com/howeyc/gopass"
	"net"
	"os"
	"strconv"
	"strings"
)

const version = "0.1"

var connectedAs string

func main() {
	daemonConnected := true

	daemonConn, identity, err := setupDaemon()
	connectedAs = identity

	if err != nil {
		if err == errDaemonNotRunning {
			daemonConnected = false
		} else if err == errUnexpectedResponse {
			fmt.Println("Sorry, an unexpected response was received from the daemon.")
			fmt.Println("Kill the daemon process and try again.")
			return
		} else {
			fmt.Println("Sorry, an issue occured while attempting to communicate with the daemon.")
			fmt.Println("    " + err.Error())
			fmt.Println("Kill the daemon process and try again.")
			return
		}
	} else {
		defer daemonConn.Close()
	}

	if len(os.Args) == 1 {
		printUsage()
		return
	} else if os.Args[1] == "remote" || os.Args[1] == "local" {
		if !daemonConnected {
			printConnectFirst()
			return
		}

		tunnelRequest(daemonConn)
	} else if os.Args[1] == "close" {
		if !daemonConnected {
			printConnectFirst()
			return
		}

		closeRequest(daemonConn)
	} else if os.Args[1] == "list" {
		if !daemonConnected {
			printConnectFirst()
			return
		}

		fmt.Println(tellDaemon(daemonConn, "list"))
	} else if os.Args[1] == "dc" || os.Args[1] == "disconnected" ||
		os.Args[1] == "logout" {
		if !daemonConnected {
			printConnectFirst()
			return
		}

		fmt.Println(tellDaemon(daemonConn, "dc"))
	} else if strings.Contains(os.Args[1], "@") {
		if daemonConnected {
			fmt.Println("Please disconnect first with")
			fmt.Println("    tport dc")
			fmt.Println("before attempting to connect to a new server.")
			return
		}

		parts := strings.Split(os.Args[1], "@")
		if len(parts) > 2 || len(parts) < 2 {
			printUsage()
		} else {
			for {
				var password string

				if len(os.Args) < 3 {
					fmt.Printf(parts[0] + "@" + parts[1] + "'s password: ")
					password = string(gopass.GetPasswd())
				} else {
					passwordData, err := hex.DecodeString(os.Args[2])
					if err != nil {
						printUsage()
						return
					}

					password = string(passwordData)
				}

				if len(password) == 0 {
					fmt.Println("Cancelled.")
					return
				}

				fmt.Println("Connecting as " + parts[0] + "@" + parts[1] + "...")

				client, err := connectToRemote(parts[0], password, parts[1])
				if err != nil {
					if strings.Contains(err.Error(), "ssh: unable to authenticate") {
						fmt.Println("Authentication failed, try again.")
						continue
					}

					fmt.Println("Failed to connect to remote server:")
					fmt.Println("    " + err.Error())
					return
				}

				newDaemon(parts[0], password, parts[1], client)
				return
			}

		}
	} else {
		printUsage()
	}
}

func closeRequest(daemonConn net.Conn) {
	if len(os.Args) < 3 {
		printUsage()
		return
	}

	if os.Args[2] == "all" {
		fmt.Println(tellDaemon(daemonConn, "close all"))
		return
	}

	if (os.Args[2] != "remote" && os.Args[2] != "local") ||
		(len(os.Args) < 4) {
		printUsage()
		return
	}

	if os.Args[3] == "all" {
		fmt.Println(tellDaemon(daemonConn, "close "+os.Args[2]+" all"))
		return
	}

	if !isNumber(os.Args[3]) {
		fmt.Println("Port must be a number.")
		return
	}

	fmt.Println(tellDaemon(daemonConn, "close "+os.Args[2]+" "+os.Args[3]))
}

func tunnelRequest(daemonConn net.Conn) {
	if len(os.Args) < 3 {
		fmt.Println("A primary port must be specified.")
		fmt.Println("    tport help for more information")
		return
	}

	if !isNumber(os.Args[2]) {
		fmt.Println("Port must be a number.")
		return
	}

	if len(os.Args) < 4 {
		fmt.Println(tellDaemon(daemonConn, "open "+os.Args[1]+" "+
			os.Args[2]))
		return
	}

	if !isNumber(os.Args[3]) {
		fmt.Println("Port must be a number.")
	}

	fmt.Println(tellDaemon(daemonConn, "open "+os.Args[1]+" "+
		os.Args[2]+" "+os.Args[3]))
}

func isNumber(str string) bool {
	if _, err := strconv.Atoi(str); err != nil {
		return false
	} else {
		return true
	}
}

func printUsage() {
	fmt.Println("tport v" + version + " by Jason Chu (1lann)")
	fmt.Println("")
	fmt.Println("First, connect with")
	fmt.Println("    tport user@host")
	fmt.Println("")
	fmt.Println("For tunnels from the local host, to the remote host.")
	fmt.Println("    tport remote remoteport [localport]")
	fmt.Println("For tunnels from the remote host, to the local host.")
	fmt.Println("    tport local localport [remoteport]")
	fmt.Println("")
	fmt.Println("    tport close remote localport")
	fmt.Println("    tport close local remoteport")
	fmt.Println("    tport close remote/local all")
	fmt.Println("    tport close all")
	fmt.Println("")
	fmt.Println("List all the open tunnels.")
	fmt.Println("    tport list")
	fmt.Println("Disconnect, close all tunnels, and quit the daemon.")
	fmt.Println("    tport dc/disconnect/logout")
	if len(connectedAs) > 0 {
		fmt.Println("")
		fmt.Println("You are currently connected as: " + connectedAs)
	}
}

func printConnectFirst() {
	fmt.Println("Sorry, you must connect first with")
	fmt.Println("    tport user@host")
	fmt.Println("before you can execute that command.")
}
