// yapi
// Copyright (c) 2014 Fatih Cetinkaya (http://github.com/cmfatih/yapi)
// For the full copyright and license information, please view the LICENSE.txt file.

// This file contains ssh client implementation.
//
// References:
//   `ClientKeyring` and `ClientPassword`: http://dave.cheney.net/?p=9

package client

import (
	"code.google.com/p/go.crypto/ssh"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

// sshClient implements a ssh client
type sshClient struct {
	id      string           // id
	name    string           // name
	kind    string           // kind of client (ssh)
	addr    string           // remote system address information
	addrF   string           // fixed remote system address information
	auth    ClientAuth       // remote system authentication information
	sshConf ssh.ClientConfig // ssh client configuration
	sshConn *ssh.ClientConn  // ssh client connection
	sshSess *ssh.Session     // ssh session
}

// ID returns the unique id of the client.
func (cliSSH *sshClient) ID() string {
	return cliSSH.id
}

// Name returns the name of the client.
func (cliSSH *sshClient) Name() string {
	return cliSSH.name
}

// Kind returns the kind of the client.
func (cliSSH *sshClient) Kind() string {
	return cliSSH.kind
}

// SetAddr sets the address information of the remote system.
// Address can be; `host` or `host:port`. Default port is `22`
func (cliSSH *sshClient) SetAddr(cliAddr string) error {

	// Check and set address
	if cliAddr == "" {
		return errors.New("missing address")
	}

	sli := strings.Split(cliAddr, ":")
	sliLen := len(sli)
	host, port := sli[0], ""

	if sliLen == 1 || (sliLen > 1 && sli[1] == "") {
		port = "22"
	} else if sliLen > 1 {
		port = sli[1]
	}
	cliAddrF := host + ":" + port

	if _, _, err := net.SplitHostPort(cliAddrF); err != nil {
		return errors.New("invalid address: " + err.Error())
	}

	cliSSH.addr = cliAddr
	cliSSH.addrF = cliAddrF

	return nil
}

// SetAuth sets the authentication information of the remote system.
func (cliSSH *sshClient) SetAuth(cliAuth ClientAuth) error {

	// Check and set auth
	ck := new(sshCK)

	// Key file
	if cliAuth.Keyfile != "" {
		if err := ck.loadPEM(cliAuth.Keyfile); err != nil {
			return errors.New("key file couldn't be read: " + cliAuth.Keyfile)
		}
	}

	if cliAuth.Username != "" || cliAuth.Password != "" || cliAuth.Keyfile != "" {
		cliSSH.sshConf = ssh.ClientConfig{
			User: cliAuth.Username,
			Auth: []ssh.ClientAuth{
				ssh.ClientAuthPassword(sshCP(cliAuth.Password)),
				ssh.ClientAuthKeyring(ck),
			},
		}
	}

	cliSSH.auth = cliAuth

	return nil
}

// Connect establishes a connection to the remote system.
func (cliSSH *sshClient) Connect() error {

	// Init vars
	var err error

	// Check the address
	if cliSSH.addrF == "" {
		return errors.New("missing address")
	}

	// Connect and init a session
	if cliSSH.sshConn, err = ssh.Dial("tcp", cliSSH.addrF, &cliSSH.sshConf); err != nil {
		return errors.New("failed to connect: " + err.Error())
	}

	if cliSSH.sshSess, err = cliSSH.sshConn.NewSession(); err != nil {
		return errors.New("failed to create session: " + err.Error())
	}

	return nil
}

// ExecCmd executes the given command on the remote system.
// It uses stdout and stderr of host.
// Be aware about return values and output! The client's stderr is different than host's stderr.
func (cliSSH *sshClient) ExecCmd(cliCmd string) (bool, error) {

	// Init vars
	var err error

	// Check vars
	if cliCmd == "" {
		return false, errors.New("missing command")
	}

	// Connection
	if err := cliSSH.Connect(); err != nil {
		return false, errors.New("connection error: " + err.Error())
	}
	defer cliSSH.sshSess.Close()

	// stdout
	stdout, err := cliSSH.sshSess.StdoutPipe()
	if err != nil {
		return false, errors.New("failed to execute (stdout): " + err.Error())
	}

	// stderr
	stderr, err := cliSSH.sshSess.StderrPipe()
	if err != nil {
		return false, errors.New("failed to execute (stderr): " + err.Error())
	}

	// Start
	if err = cliSSH.sshSess.Start(cliCmd); err != nil {
		return false, errors.New("failed to execute: " + err.Error())
	}

	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	return true, cliSSH.sshSess.Wait()
}

// sshCK implements the ClientKeyring interface.
type sshCK struct {
	keys []ssh.Signer
}

// Key returns the public key.
func (k *sshCK) Key(i int) (ssh.PublicKey, error) {
	if i < 0 || i >= len(k.keys) {
		return nil, nil
	}

	return k.keys[i].PublicKey(), nil
}

// Sign returns the signature.
func (k *sshCK) Sign(i int, rand io.Reader, data []byte) (sig []byte, err error) {
	return k.keys[i].Sign(rand, data)
}

// add appends the given key to the keys.
func (k *sshCK) add(key ssh.Signer) {
	k.keys = append(k.keys, key)
}

// loadPEM loads and parses private key by the given file path.
func (k *sshCK) loadPEM(file string) error {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	key, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		return err
	}
	k.add(key)
	return nil
}

// sshCP implements the ClientPassword interface.
type sshCP string

// Password returns the password.
func (p sshCP) Password(user string) (string, error) {
	return string(p), nil
}
