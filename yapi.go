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
	"runtime"
	"strings"
)

const (
	YAPI_VERSION = "HEAD^0.3.0" // app version
)

var (
	gvCLEC        string    // command line escape char
	gvPipeConf    pipe.Conf // pipe config
	flHelp        bool      // help flag
	flVersion     bool      // version flag
	flPipeConf    string    // pipe config flag
	flClientName  string    // client name flag
	flClientGroup string    // client group flag
	flClientCmd   string    // client command flag
	flClientCEM   string    // client command execution method
)

func main() {

	// Check flags
	flag.Parse()

	if flVersion == true {
		cmdVer()
		return
	} else if flHelp == true || flClientCmd == "" {
		cmdUsage()
		return
	}

	// Init the pipe config
	err := gvPipeConf.Load(flPipeConf)
	if err != nil {
		fmt.Printf("Failed to load pipe configuration: %s", err)
		return
	} else if gvPipeConf.IsLoaded() == false {
		fmt.Printf("Failed to load pipe configuration. Please use [-pc FILE] option.")
		return
	}

	// Determine the clients
	cliNames := flagMultiParser(flClientName, ",")
	cliGroups := flagMultiParser(flClientGroup, ",")
	if cliNames == nil && cliGroups == nil {
		if _, cliDefName := gvPipeConf.ClientDef(); cliDefName != "" {
			cliNames = append(cliNames, cliDefName)
		}
	} else if cliGroups != nil {
		for _, val := range cliGroups {
			cliList := client.ByGroupName(val)
			for key, _ := range cliList {
				cliNames = append(cliNames, cliList[key].Name())
			}
		}
	}
	if cliNames == nil {
		fmt.Print("Failed to determine a client. Please use [-cn CLIENTNAME] or [-cg GROUPNAME] option.")
		return
	}

	// Execute the client command
	if flClientCmd != "" {
		cliCCEM := flagSymbolParser(flClientCEM)
		if cliCCEM != "serial" && cliCCEM != "parallel" {
			fmt.Println("Invalid client command execution method. Please use [-ccem serial] or [-ccem parallel]")
			return
		}

		ccew, err := worker.New("cce")
		if err != nil {
			fmt.Printf("Failed to execute the command: %s\n", err)
			return
		}
		if err := ccew.SetOptions(
			worker.WorkerOptions{
				Putty: worker.CCEOptions{Clients: cliNames, Cmd: flClientCmd, CmdErrPrint: true, Method: cliCCEM},
			},
		); err != nil {
			fmt.Printf("Failed to execute the command: %s\n", err)
			return
		}
		if err := ccew.Start(); err != nil {
			fmt.Printf("Failed to execute the command: %s\n", err)
			return
		}
	}
}

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

// cmdUsage displays the usage of the app
func cmdUsage() {
	fmt.Print("Usage: yapi [OPTION]...\n\n")
	fmt.Printf("yapi - Yet Another Pipe Implementation - v%s\n", YAPI_VERSION)

	fmt.Printf("\nOptions:")
	//flag.PrintDefaults()
	/*
		flag.VisitAll(func(flg *flag.Flag) {
			defVal := ""
			if flg.DefValue != "" && flg.DefValue != "false" {
				defVal = "(default: " + flg.DefValue + ")"
			}
			fmt.Printf("  -%-10s : %s %s\n", flg.Name, flg.Usage, defVal)
		})
	*/
	fmt.Print(`
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
  `)

	fmt.Printf("\nExamples:")
	fmt.Print(`
  yapi -cc ls
  yapi -pc /path/pipe.json -cc "tail -F /var/log/syslog" -ccem parallel
  yapi -cc "top -b -n 1" | grep ssh
  yapi -cn client1 -cc "ps aux" | yapi -cn client2 -cc "wc -l"
  yapi -cc hostname -cn "client1,client2" -ccem parallel
  yapi -cc hostname -cg group1 -ccem parallel
  `)

	fmt.Printf("\nPlease report issues to https://github.com/cmfatih/yapi/issues\n")
}

// cmdVer displays the version information of the app
func cmdVer() {
	fmt.Printf("yapi version %s\n", YAPI_VERSION)
}

// flagMultiParser parses multiple flag values.
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

// flagSymbolParser parses symbol flag values.
func flagSymbolParser(flagVal string) string {
	if flagVal == "~" {
		return "serial"
	} else if flagVal == "//" {
		return "parallel"
	}

	return flagVal
}
