// yapi
// Copyright (c) 2014 Fatih Cetinkaya (http://github.com/cmfatih/yapi)
// For the full copyright and license information, please view the LICENSE.txt file.

// Package pipe provides pipe related functions.
package pipe

import (
	"encoding/json"
	"errors"
	"github.com/cmfatih/yapi/client"
	"io/ioutil"
	"os"
	"strconv"
)

// Conf implements the pipe configuration.
type Conf struct {
	isLoaded bool
	filePath string

	Clients       []confClient `json:"clients"`
	clientDefID   string
	clientDefName string
}

type confClient struct {
	ID        string
	Name      string         `json:"name"`
	Kind      string         `json:"kind"`
	Address   string         `json:"address"`
	Auth      confClientAuth `json:"auth"`
	IsDefault bool           `json:"isDefault"`
}

type confClientAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Keyfile  string `json:"keyfile"`
}

// IsLoaded returns whether the configuration is loaded or not.
func (conf *Conf) IsLoaded() bool {
	return conf.isLoaded
}

// Load loads the configuration file by the given file path.
// Default file path is `pipe.json`.
func (conf *Conf) Load(filePath string) error {

	// Init vars
	isDefFile := false
	conf.isLoaded = false

	// Check the file path
	if filePath == "" {
		filePath = "pipe.json" // default
		isDefFile = true
	}

	// Check the file path whether exists and readable or not
	fInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) == false {
			return errors.New("file is not readable: " + filePath)
		} else if os.IsNotExist(err) == true {
			if isDefFile == false {
				return errors.New("file is not found: " + filePath)
			} else if os.IsNotExist(err) == true && isDefFile == true {
				// Do not return error for non-exists default file
				return nil
			}
		}
	} else if fInfo == nil || fInfo.IsDir() == true {
		// Check the file whether a directory or not
		return errors.New("invalid file: " + filePath)
	}

	// Load the file
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return errors.New("failed to read: " + err.Error())
	}

	// Parse the content
	if err := json.Unmarshal(contents, &conf); err != nil {
		return errors.New("failed to parse: " + err.Error())
	}

	// Init clients
	defCliID := ""
	defCliName := ""

	for cliInd, cliConf := range conf.Clients {

		cli, err := client.New(cliConf.Kind, cliConf.Name)
		if err != nil {
			return errors.New("failed to create the client (index: " + strconv.Itoa(cliInd) + ", name: " + cliConf.Name + "): " + err.Error())
		}

		if err := cli.SetAuth(client.ClientAuth{
			Username: cliConf.Auth.Username,
			Password: cliConf.Auth.Password,
			Keyfile:  cliConf.Auth.Keyfile,
		}); err != nil {
			return errors.New("error on client auth (index: " + strconv.Itoa(cliInd) + ", name: " + cliConf.Name + "): " + err.Error())
		}

		if err := cli.SetAddr(cliConf.Address); err != nil {
			return errors.New("error on client address (index: " + strconv.Itoa(cliInd) + ", name: " + cliConf.Name + "): " + err.Error())
		}

		// Default client
		if cliConf.IsDefault == true {
			defCliID = cli.ID()
			defCliName = cli.Name()
		}

		// Set the ID
		conf.Clients[cliInd].ID = cli.ID()
	}

	conf.isLoaded = true
	conf.filePath = filePath
	conf.clientDefID = defCliID
	conf.clientDefName = defCliName

	return nil
}

// ClientDef returns the id and name of the default client if any.
func (conf *Conf) ClientDef() (string, string) {
	return conf.clientDefID, conf.clientDefName
}
