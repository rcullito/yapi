// yapi
// Copyright (c) 2014 Fatih Cetinkaya (http://github.com/cmfatih/yapi)
// For the full copyright and license information, please view the LICENSE.txt file.

// yapi - Yet Another Pipe Implementation
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cmfatih/yapi/client"
	"github.com/cmfatih/yapi/pipe"
	"github.com/cmfatih/yapi/worker"
	"os"
	"runtime"
	"strings"
)

const (
	YAPI_VERSION = "0.3.1^HEAD" // app version
)

var (
	gvCLEC      string    // command line escape char
	gvHOME      string    // User home directory
	gvPipeConf  pipe.Conf // pipe config
	gvCliNames  []string  // client names
	gvCliGroups []string  // client groups
	flHelp      bool      // help flag
	flVersion   bool      // version flag
	flPipeConf  string    // pipe config flag
	flCliName   string    // client name flag
	flCliGroup  string    // client group flag
	flCliCmd    string    // client command flag
	flCliCEM    string    // client command execution method
	flSSH       string    // simple ssh client
)

func init() {

	// Init vars
	if runtime.GOOS == "windows" {
		gvCLEC = "^"
		gvHOME = os.Getenv("USERPROFILE")
	} else {
		gvCLEC = "\\"
		gvHOME = os.Getenv("HOME")
	}

	// Init flags
	flag.BoolVar(&flHelp, "help", false, "Display help and exit.")
	flag.BoolVar(&flHelp, "h", false, "Display help and exit.")
	flag.BoolVar(&flVersion, "version", false, "Display version information and exit.")
	flag.BoolVar(&flVersion, "v", false, "Display version information and exit.")

	flag.StringVar(&flCliCmd, "cc", "", "Client command that will be executed.")
	flag.StringVar(&flCliName, "cn", "", "Client name(s) those will be connected.")
	flag.StringVar(&flCliGroup, "cg", "", "Client group name(s) those will be connected.")
	flag.StringVar(&flCliCEM, "ccem", "serial", "Execution method for client command. Default; serial")
	flag.StringVar(&flPipeConf, "pc", "", "Pipe configuration file. Default; pipe.json")

	flag.StringVar(&flSSH, "ssh", "", "Simple SSH client command execution.")
}

func main() {

	// Init flags
	flag.Parse()

	flagVer(flVersion) // version
	flagHelp(flHelp)   // help

	if flSSH != "" {
		// Simple SSH CCE
		flagSSH(flSSH, flCliCmd, flCliCEM)
	} else if flCliCmd != "" {
		// Client command execution
		flagPC(flPipeConf)                     // pipe config
		flagCNG(flCliName, flCliGroup)         // client names and groups
		flagCC(flCliCmd, flCliCEM, gvCliNames) // client command
	}

	flagHelp(true) // Default
}

// flagVer displays the version information and exit.
func flagVer(ver bool) {

	if ver != true {
		return
	}

	fmt.Printf("yapi version %s\n", YAPI_VERSION)

	flagExit()
}

// flagHelp displays the help and exit.
func flagHelp(help bool) {

	if help != true {
		return
	}

	fmt.Print("Usage: yapi [OPTION]...\n\n")
	fmt.Printf("yapi - Yet Another Pipe Implementation - v%s\n", YAPI_VERSION)
	fmt.Print(`
  Options:
    -cc       : Client command that will be executed.
    -cn       : Client name(s) those will be connected.
                Use comma (,) for multi-client.
    -cg       : Client group name(s) those will be connected. 
                Use comma (,) for multi-group.
    -ccem     : Execution method for client command. Default; serial
                Possible values; serial (~), parallel (//)

    -pc       : Pipe configuration file. Default; pipe.json

    -ssh      : Simple SSH client command execution.
                It uses the current/given username and HOME/.ssh/id_rsa 
                for the private key file.
                Syntax: [user@]host[:22]

    -help     : Display help and exit.
    -h
    -version  : Display version information and exit.
    -v

  Examples:
    yapi -cc ls
    yapi -cc "top -b -n 1" | grep ssh
    yapi -cc "tail -F /var/log/syslog" -ccem parallel
    yapi -cc hostname -cn "client1,client2" -ccem parallel
    yapi -cc hostname -cg group1 -ccem parallel
    yapi -cc "ps aux" -cn client1 | yapi -cc "wc -l" -cn client2

    yapi -ssh localhost -cc ls
    yapi -ssh user@localhost:22 -cc ls
    yapi -ssh host1,host2 -cc ls -ccem parallel


  Please report issues to https://github.com/cmfatih/yapi/issues

  `)

	flagExit()
}

// flagPC loads pipe config.
func flagPC(pcFile string) {

	if err := gvPipeConf.Load(pcFile, pipe.LoadOpt{CliInit: true}); err != nil {
		fmt.Printf("Error due pipe configuration: %s\n", err)
		flagExit()
	}

	return
}

// flagCNG sets the client names and groups.
func flagCNG(cliName, cliGroup string) {

	gvCliNames = flagMultiParser(cliName, ",")
	gvCliGroups = flagMultiParser(cliGroup, ",")

	if gvCliGroups != nil {
		gvCliNames = []string{} // Groups overwrites names
		for _, val := range gvCliGroups {
			cl := client.ByGroupName(val)
			for key, _ := range cl {
				gvCliNames = append(gvCliNames, cl[key].Name())
			}
		}
	}

	return
}

// flagCC executes the client command if any.
func flagCC(cliCmd, cliCmdEM string, cliNames []string) {

	if cliNames == nil {
		if _, name := gvPipeConf.CliDef(); name != "" {
			cliNames = append(cliNames, name) // Default client
		}
	}

	// Create a new worker
	ccew, err := worker.New("cce")
	if err != nil {
		fmt.Printf("Failed to execute the command: %s\n", err)
		flagExit()
	}

	// Set the options
	if err := ccew.SetOptions(
		worker.WorkerOptions{
			Putty: worker.CCEOptions{
				Clients:     cliNames,
				Cmd:         cliCmd,
				CmdErrPrint: true,
				Method:      flagSymbolParser(cliCmdEM),
			},
		},
	); err != nil {
		fmt.Printf("Failed to execute the command: %s\n", err)
		flagExit()
	}

	// Start the worker
	if err := ccew.Start(); err != nil {
		fmt.Printf("Failed to execute the command: %s\n", err)
		flagExit()
	}

	flagExit()
}

// flagSSH executes the given command via ssh client.
func flagSSH(sshOpt, cliCmd, cliCmdEM string) {

	// Init vars
	cliAddrs := flagMultiParser(sshOpt, ",")
	cliConfs := []string{}
	cliNames := []string{}

	for key, val := range cliAddrs {
		name := fmt.Sprintf("ssh_%d", key)
		addr := val
		authUN := "" // default; current user
		authKf := gvHOME + "/.ssh/id_rsa"

		if strings.Contains(val, "@") == true {
			spl := strings.SplitN(val, "@", 2)
			authUN = spl[0]
			if len(spl) > 1 {
				addr = spl[1]
			}
		}

		// Prepare JSON content
		jc := map[string]interface{}{
			"name":      name,
			"kind":      "ssh",
			"isDefault": true,
			"address":   addr,
			"auth": map[string]interface{}{
				"username": authUN,
				"keyfile":  authKf,
			},
		}
		if buf, err := json.Marshal(jc); err != nil {
			fmt.Printf("Unexpected error: %s\n", err)
			flagExit()
		} else {
			cliConfs = append(cliConfs, string(buf))
			cliNames = append(cliNames, name)
		}
	}

	jsonCont := "{\"Clients\":[" + strings.Join(cliConfs, ",") + "]}"

	if err := gvPipeConf.LoadJSON(jsonCont, pipe.LoadOpt{CliInit: true}); err != nil {
		fmt.Printf("Error due pipe configuration: %s\n", err)
		flagExit()
	}

	flagCC(cliCmd, cliCmdEM, cliNames)

	flagExit()
}

// flagMultiParser parses multiple flag value.
func flagMultiParser(flagVal, valSep string) []string {

	if flagVal != "" {
		var flagVals []string
		founds := make(map[string]bool)
		spl := strings.Split(flagVal, valSep)
		for _, val := range spl {
			val = strings.TrimSpace(val)
			if val != "" && founds[val] == false {
				flagVals = append(flagVals, val)
				founds[val] = true
			}
		}

		return flagVals
	}

	return nil
}

// flagSymbolParser parses symbol flag value.
func flagSymbolParser(flagVal string) string {

	if flagVal == "~" {
		return "serial"
	} else if flagVal == "//" {
		return "parallel"
	}

	return flagVal
}

// flagExit
func flagExit() {
	os.Exit(0)
}
