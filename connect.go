package main

import (
	"golang.org/x/crypto/ssh"
	"strings"
)

func connectToRemote(username, password, host string) (*ssh.Client, error) {
	sshConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
	}

	if !strings.Contains(host, ":") {
		host = host + ":22"
	}

	client, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return client, err
	}

	return client, nil
}
