// yapi
// Copyright (c) 2014 Fatih Cetinkaya (http://github.com/cmfatih/yapi)
// For the full copyright and license information, please view the LICENSE.txt file.

// yapi - Yet Another Pipe Implementation
package main

import (
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
	gvCLEC        string    // command line escape char
	gvPipeConf    pipe.Conf // pipe config
	gvClientNames []string  // client names
	flHelp        bool      // help flag
	flVersion     bool      // version flag
	flPipeConf    string    // pipe config flag
	flClientName  string    // client name flag
	flClientGroup string    // client group flag
	flClientCmd   string    // client command flag
	flClientCEM   string    // client command execution method
)

func init() {

	// Init vars
	if runtime.GOOS == "windows" {
		gvCLEC = "^"
	} else {
		gvCLEC = "\\"
	}

	// Init flags
	flag.BoolVar(&flHelp, "help", false, "Display help and exit.")
	flag.BoolVar(&flHelp, "h", false, "Display help and exit.")
	flag.BoolVar(&flVersion, "version", false, "Display version information and exit.")
	flag.BoolVar(&flVersion, "v", false, "Display version information and exit.")

	flag.StringVar(&flClientCmd, "cc", "", "Client command that will be executed.")
	flag.StringVar(&flClientName, "cn", "", "Client name(s) those will be connected.")
	flag.StringVar(&flClientGroup, "cg", "", "Client group name(s) those will be connected.")
	flag.StringVar(&flClientCEM, "ccem", "serial", "Execution method for client command. Default; serial")
	flag.StringVar(&flPipeConf, "pc", "", "Pipe configuration file. Default; pipe.json")
}

func main() {

	// Init flags
	flag.Parse()

	flagVersionParser(flVersion)                          // version
	flagHelpParser(flHelp)                                // help
	flagPCParser(flPipeConf)                              // pipe config
	flagCNGParser(flClientName, flClientGroup)            // client names
	flagCCParser(flClientCmd, flClientCEM, gvClientNames) // client command
	flagHelpParser(true)                                  // Default

	return
}

// flagVersionParser displays the version information and exit.
func flagVersionParser(ver bool) {

	if ver == true {
		fmt.Printf("yapi version %s\n", YAPI_VERSION)
		os.Exit(0)
	}

	return
}

// flagHelpParser displays the usage and exit.
func flagHelpParser(help bool) {

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


  Please report issues to https://github.com/cmfatih/yapi/issues

  `)
	os.Exit(0)

	return
}

// flagCCParser parses `-cc` flag.
func flagCCParser(cc, ccem string, clis []string) {

	if cc == "" {
		return
	}

	// Execute the client command
	ccew, err := worker.New("cce")
	if err != nil {
		fmt.Printf("Failed to execute the command: %s\n", err)
		return
	}
	if err := ccew.SetOptions(
		worker.WorkerOptions{
			Putty: worker.CCEOptions{
				Clients:     clis,
				Cmd:         cc,
				CmdErrPrint: true,
				Method:      flagSymbolParser(ccem),
			},
		},
	); err != nil {
		fmt.Printf("Failed to execute the command: %s\n", err)
		os.Exit(0)
	}
	if err := ccew.Start(); err != nil {
		fmt.Printf("Failed to execute the command: %s\n", err)
		os.Exit(0)
	}

	os.Exit(0)

	return
}

// flagPCParser parses `-pc` flag.
func flagPCParser(pc string) {

	err := gvPipeConf.Load(pc)
	if err != nil {
		fmt.Printf("Failed to load pipe configuration: %s", err)
		return
	} else if gvPipeConf.IsLoaded() == false && pc != "" {
		fmt.Printf("Failed to load pipe configuration. Please use [-pc FILE] option.")
		return
	}

	return
}

// flagCNGParser parses `-cn` and `-cg` flags and set gvClientNames var.
func flagCNGParser(cn, cg string) {

	gvClientNames = flagMultiParser(cn, ",")
	cliGroups := flagMultiParser(cg, ",")
	if gvClientNames == nil && cliGroups == nil {
		if _, cliDefName := gvPipeConf.ClientDef(); cliDefName != "" {
			gvClientNames = append(gvClientNames, cliDefName)
		}
	} else if cliGroups != nil {
		for _, val := range cliGroups {
			cliList := client.ByGroupName(val)
			for key, _ := range cliList {
				gvClientNames = append(gvClientNames, cliList[key].Name())
			}
		}
	}

	return
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
