// yapi
// Copyright (c) 2014 Fatih Cetinkaya (http://github.com/cmfatih/yapi)
// For the full copyright and license information, please view the LICENSE.txt file.

// This file contains docker client implementation.
//
// References:
//   dockerclient 	: https://github.com/samalba/dockerclient
//   Authentication :
//   	https://github.com/dotcloud/docker/pull/3068
//   	http://docs.docker.io/en/latest/use/basics/#bind-docker-to-another-host-port-or-a-unix-socket
//
// Todo:
// 	Authentication:
// 		Currently there is no authentication for the Docker remote API. Keep check it and
// 		implement when it is ready.

package client

import (
	"errors"
	"fmt"
	dcli "github.com/samalba/dockerclient"
	"net/url"
)

var _ = fmt.Println // for debug

// dockerClient implements a docker client
type dockerClient struct {
	id        string             // id
	name      string             // name
	kind      string             // kind of client (docker)
	addr      string             // remote system address information
	addrF     string             // fixed remote system address information
	auth      ClientAuth         // remote system authentication information
	dockerCli *dcli.DockerClient // docker client
}

// ID returns the unique id of the client.
func (cliDocker *dockerClient) ID() string {
	return cliDocker.id
}

// Name returns the name of the client.
func (cliDocker *dockerClient) Name() string {
	return cliDocker.name
}

// Kind returns the kind of the client.
func (cliDocker *dockerClient) Kind() string {
	return cliDocker.kind
}

// SetAddr sets the address information of the remote system.
// Address can be; `unix://path` or `host:port`.
func (cliDocker *dockerClient) SetAddr(cliAddr string) error {

	// Check and set address
	if cliAddr == "" {
		return errors.New("missing address")
	}

	up, err := url.Parse(cliAddr)
	if err != nil {
		return errors.New("invalid address: " + err.Error())
	} else if up.Scheme != "unix" && up.Scheme != "http" {
		return errors.New("for scheme use unix:// or http://")
	}

	cliAddrF := up.Scheme + "://" + up.Path

	cliDocker.addr = cliAddr
	cliDocker.addrF = cliAddrF

	return nil
}

// SetAuth sets the authentication information of the remote system.
func (cliDocker *dockerClient) SetAuth(cliAuth ClientAuth) error {

	// Check and set auth
	cliDocker.auth = cliAuth

	return nil
}

// Connect establishes a connection to the remote system.
func (cliDocker *dockerClient) Connect() error {

	// Init vars
	var err error

	// Check the address
	if cliDocker.addrF == "" {
		return errors.New("missing address")
	}

	// Connect
	if cliDocker.dockerCli, err = dcli.NewDockerClient(cliDocker.addrF); err != nil {
		return errors.New("failed to connect: " + err.Error())
	}

	return nil
}

// ExecCmd executes the given command on the remote system.
// It uses stdout and stderr of host.
// Be aware about return values and output! The client's stderr is different than host's stderr.
func (cliDocker *dockerClient) ExecCmd(cliCmd string) (bool, error) {

	// Init vars
	//var err error

	// Check vars
	if cliCmd == "" {
		return false, errors.New("missing command")
	}

	// Connection
	if err := cliDocker.Connect(); err != nil {
		return false, errors.New("connection error: " + err.Error())
	}
	// MonitorEvents is not in use but just in case...
	defer cliDocker.dockerCli.StopAllMonitorEvents()

	//fmt.Println(cliDocker)

	return false, errors.New("docker client implementation is still under development...")

	return true, nil
}
