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
	"github.com/cmfatih/yapi/stdin"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"runtime"
	"strings"
)

// sshClient implements a ssh client
type sshClient struct {
	id      string           // id
	name    string           // name
	groups  []string         // groups
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

// Groups returns the groups of the client.
func (cliSSH *sshClient) Groups() []string {
	return cliSSH.groups
}

// SetGroups sets the groups of client.
func (cliSSH *sshClient) SetGroups(cliGroups []string) error {

	// Check the group names
	for _, val := range cliGroups {
		if err := nameCheck(val, "word"); err != nil {
			return errors.New("invalid group name (" + val + "), " + err.Error())
		}
	}

	cliSSH.groups = cliGroups

	return nil
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

	// Determine the username
	// For SSH protocol username is required. So try to determine it if possible.
	if cliAuth.Username == "" {

		// WARN: It doesn't work on OSX if the cross compiler used

		if u, err := user.Current(); err == nil && u != nil {
			if u.Username != "" {
				if runtime.GOOS == "windows" {
					sli := strings.Split(u.Username, "\\")
					sliLen := len(sli)
					if sliLen > 0 {
						cliAuth.Username = sli[sliLen-1]
					}
				} else {
					cliAuth.Username = u.Username
				}
			}
		}
	}

	// Key file
	if cliAuth.Keyfile != "" {
		if err := ck.loadPEM(cliAuth.Keyfile); err != nil {
			return errors.New("key file couldn't be read: " + cliAuth.Keyfile + " - " + err.Error())
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

	// client stdin
	cliStdin, err := cliSSH.sshSess.StdinPipe()
	if err != nil {
		return false, errors.New("failed to execute (stdin): " + err.Error())
	}

	// client stdout
	cliStdout, err := cliSSH.sshSess.StdoutPipe()
	if err != nil {
		return false, errors.New("failed to execute (stdout): " + err.Error())
	}

	// client stderr
	cliStderr, err := cliSSH.sshSess.StderrPipe()
	if err != nil {
		return false, errors.New("failed to execute (stderr): " + err.Error())
	}

	// Start
	if err = cliSSH.sshSess.Start(cliCmd); err != nil {
		return false, errors.New("failed to execute: " + err.Error())
	}

	if stdin.StdinHasPipe() == true {
		_, err := io.Copy(cliStdin, stdin.StdinReader())
		if err != nil {
			return false, errors.New("failed to copy stdin: " + err.Error())
		}
		cliStdin.Close()
	} else {
		cliStdin.Close()
	}

	go io.Copy(os.Stdout, cliStdout)
	go io.Copy(os.Stderr, cliStderr)

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
