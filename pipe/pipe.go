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
	isLoaded    bool
	isCliInited bool
	filePath    string

	Clients       []confClient `json:"clients"`
	clientDefID   string
	clientDefName string
}

type confClient struct {
	ID        string
	Name      string         `json:"name"`
	Groups    []string       `json:"groups"`
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

type LoadOpt struct {
	CliInit bool
}

// IsLoaded returns whether the configuration is loaded or not.
func (conf *Conf) IsLoaded() bool {
	return conf.isLoaded
}

// IsCliInited returns whether the client configuration is initialized or not.
func (conf *Conf) IsCliInited() bool {
	return conf.isCliInited
}

// Load loads the configuration by the given file path.
// Default file path is `pipe.json`.
func (conf *Conf) Load(filePath string, opt LoadOpt) error {

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

	conf.isLoaded = true
	conf.filePath = filePath

	if opt.CliInit == true {
		if err := conf.CliInit(); err != nil {
			return errors.New("failed to initialize clients: " + err.Error())
		}
	}

	return nil
}

// Load loads the configuration by the given JSON content.
func (conf *Conf) LoadJSON(jsonCont string, opt LoadOpt) error {

	// Init vars
	conf.isLoaded = false

	// Check the content
	if jsonCont == "" {
		return errors.New("invalid JSON content")
	}

	// Parse the content
	if err := json.Unmarshal([]byte(jsonCont), &conf); err != nil {
		return errors.New("failed to parse: " + err.Error())
	}

	conf.isLoaded = true

	if opt.CliInit == true {
		if err := conf.CliInit(); err != nil {
			return errors.New("failed to initialize clients: " + err.Error())
		}
	}

	return nil
}

// CliInit initialize the clients.
func (conf *Conf) CliInit() error {

	// Init vars
	defCliID := ""
	defCliName := ""

	for cliInd, cliConf := range conf.Clients {

		// Create client
		cli, err := client.New(cliConf.Kind, cliConf.Name)
		if err != nil {
			return errors.New("failed to create the client (index: " + strconv.Itoa(cliInd) + ", name: " + cliConf.Name + "): " + err.Error())
		}

		// Set the ID
		conf.Clients[cliInd].ID = cli.ID()

		// Set groups
		if err := cli.SetGroups(cliConf.Groups); err != nil {
			return errors.New("error on client groups (index: " + strconv.Itoa(cliInd) + ", name: " + cliConf.Name + "): " + err.Error())
		}

		// Set address
		if err := cli.SetAddr(cliConf.Address); err != nil {
			return errors.New("error on client address (index: " + strconv.Itoa(cliInd) + ", name: " + cliConf.Name + "): " + err.Error())
		}

		// Set auth
		if err := cli.SetAuth(client.ClientAuth{
			Username: cliConf.Auth.Username,
			Password: cliConf.Auth.Password,
			Keyfile:  cliConf.Auth.Keyfile,
		}); err != nil {
			return errors.New("error on client auth (index: " + strconv.Itoa(cliInd) + ", name: " + cliConf.Name + "): " + err.Error())
		}

		// Default client
		if cliConf.IsDefault == true {
			defCliID = cli.ID()
			defCliName = cli.Name()
		}
	}

	conf.clientDefID = defCliID
	conf.clientDefName = defCliName

	conf.isCliInited = true

	return nil
}

// CliDef returns the id and name of the default client if any.
func (conf *Conf) CliDef() (string, string) {
	return conf.clientDefID, conf.clientDefName
}
