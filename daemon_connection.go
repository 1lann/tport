package main

import (
	"errors"
	"net"
	"strings"
	"time"
)

const timeout = time.Second * 5

var errUnexpectedResponse = errors.New("tport: unexpected response")
var errDaemonNotRunning = errors.New("tport: daemon not running")

func readResponse(conn net.Conn) (string, error) {
	conn.SetReadDeadline(time.Now().Add(timeout))

	line := ""
	for {
		data := make([]byte, 1000)
		n, err := conn.Read(data)

		if n > 0 {
			line = line + string(data[:n])

			if data[n-1] == '\r' {
				return line[:len(line)-1], nil
			}
		}

		if err != nil {
			return line, err
		}
	}

}

func setupDaemon() (net.Conn, string, error) {
	// Check for existing tport
	var daemonConn net.Conn

	var err error

	daemonConn, err = net.Dial("tcp", "127.0.0.1:"+daemonPort)
	if err != nil {
		// Daemon not running
		return daemonConn, "", errDaemonNotRunning
	}

	daemonConn.Write([]byte("tport hello\r"))
	resp, err := readResponse(daemonConn)
	if err != nil {
		daemonConn.Close()
		return daemonConn, "", err
	}

	if resp[:6] != "tport " {
		daemonConn.Close()
		return daemonConn, "", errUnexpectedResponse
	}

	return daemonConn, resp[6:len(resp)], nil
}

func tellDaemon(conn net.Conn, msg string) string {
	_, err := conn.Write([]byte("tport " + msg + "\r"))
	if err != nil {
		return "Request failed: An error occured while attempting to write.\n" +
			"    " + err.Error()
	}

	resp, err := readResponse(conn)
	if err != nil {
		if strings.Contains(err.Error(), "i/o timeout") {
			return "Sorry, the daemon did not respond in time. It may be frozen."
		} else {
			return "Request failed: An error occured while waiting for a response.\n" +
				"    " + err.Error()
		}
	} else {
		if resp[len(resp)-1] == '\n' {
			return resp[:len(resp)-1]
		}

		return resp
	}
}
