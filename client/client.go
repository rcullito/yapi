// yapi
// Copyright (c) 2014 Fatih Cetinkaya (http://github.com/cmfatih/yapi)
// For the full copyright and license information, please view the LICENSE.txt file.

// Package client provides client related functions.
package client

import (
	"code.google.com/p/go-uuid/uuid"
	"errors"
	"regexp"
)

var (
	clients     = map[string]Client{}
	clientKinds = map[string]bool{"ssh": true, "docker": true}
	clientNames = map[string]string{}
)

// Client is the interface that must be implemented by clients.
type Client interface {

	// ID returns the unique id of the client.
	ID() string

	// Name returns the name of the client.
	Name() string

	// Kind returns the kind of the client.
	Kind() string

	// SetAddr sets the address information of the remote system.
	SetAddr(cliAddr string) error

	// SetAuth sets the authentication information of the remote system.
	SetAuth(cliAuth ClientAuth) error

	// Connect establishes a connection to the remote system.
	Connect() error

	// ExecCmd executes the given command on the remote system.
	ExecCmd(cliCmd string) (bool, error)
}

// ClientAuth implements authentication info.
// Username, Password and Keyfile are universal for authentication.
// Consider other methods (ssh-agent, db, etc.) at the future.
type ClientAuth struct {
	Username string
	Password string
	Keyfile  string
}

// New returns a new client with the given kind and name.
func New(cliKind, cliName string) (Client, error) {

	// Check vars
	if cliName == "" {
		return nil, errors.New("invalid client name")
	} else if cliKind == "" || clientKinds[cliKind] != true {
		return nil, errors.New("invalid kind (" + cliKind + ")")
	}

	// Check the name
	if r, err := regexp.Compile(`^[[:word:]]+$`); err != nil || r.MatchString(cliName) != true {
		return nil, errors.New("invalid client name, only word characters ([A-Za-z0-9_]) are allowed")
	}

	// Init client
	cliID := uuid.New()

	if cliKind == "ssh" {
		cli := sshClient{
			id:   cliID,
			name: cliName,
			kind: cliKind,
		}

		// Add to the lists
		clients[cliID] = &cli
		clientNames[cliName] = cliID

		return &cli, nil

	} else if cliKind == "docker" {
		cli := dockerClient{
			id:   cliID,
			name: cliName,
			kind: cliKind,
		}

		// Add to the lists
		clients[cliID] = &cli
		clientNames[cliName] = cliID

		return &cli, nil
	}

	return nil, errors.New("unexpected error! (client.New)")
}

// ByID returns the client by the given id.
func ByID(cliID string) (Client, error) {
	if cliID == "" || clients[cliID] == nil {
		return nil, errors.New("client is not found: " + cliID)
	}

	return clients[cliID], nil
}

// ByName returns the client by the given name.
func ByName(cliName string) (Client, error) {
	if cliName == "" || clientNames[cliName] == "" || clients[clientNames[cliName]] == nil {
		return nil, errors.New("client is not found: " + cliName)
	}

	return clients[clientNames[cliName]], nil
}

// ExecCmd executes the given command on the remote system by the given client name.
func ExecCmd(cliCmd, cliName string) error {

	// Get the client
	cli, err := ByName(cliName)
	if err != nil {
		return err
	}

	// Execute the command
	cliCO, err := cli.ExecCmd(cliCmd)
	if err != nil {
		if cliCO == false {
			// error by host
			return err
		} else {
			// error by client
			//return err
		}
	}

	return nil
}
