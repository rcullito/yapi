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
	"runtime"
	"strings"
)

const (
	YAPI_VERSION = "0.2.4" // app version
)

var (
	gvOS         string    // OS
	gvCLEC       string    // command line escape char
	gvPipeConf   pipe.Conf // pipe config
	flHelp       bool      // help flag
	flVersion    bool      // version flag
	flPipeConf   string    // pipe config flag
	flClientName string    // client name flag
	flClientCmd  string    // client command flag
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
		fmt.Printf("Failed to load pipe configuration. Please use [-pc=FILE] option.")
		return
	}

	// Init the clients
	cliNames := flagCN(flClientName)
	if cliNames == nil {
		if _, cliDefName := gvPipeConf.ClientDef(); cliDefName != "" {
			cliNames = append(cliNames, cliDefName)
		}
	}
	if cliNames == nil {
		fmt.Print("Failed to determine a client. Please use [-cn=CLIENTNAME] option.")
		return
	}

	// Execute the command
	for _, cliName := range cliNames {
		if err := client.ExecCmd(flClientCmd, cliName); err != nil {
			fmt.Printf("Failed to execute the command: %s", err)
			return
		}
	}
}

func init() {

	// Init vars
	gvOS = runtime.GOOS
	if gvOS == "windows" {
		gvCLEC = "^"
	} else {
		gvCLEC = "\\"
	}

	// Init flags
	flag.BoolVar(&flHelp, "help", false, "Display this help and exit")
	flag.BoolVar(&flHelp, "h", false, "Display this help and exit")

	flag.BoolVar(&flVersion, "version", false, "Display version information and exit")
	flag.BoolVar(&flVersion, "v", false, "Display version information and exit")

	flag.StringVar(&flPipeConf, "pc", "", "Pipe configuration file")
	flag.StringVar(&flClientName, "cn", "", "Client name")
	flag.StringVar(&flClientCmd, "cc", "", "Client command")
}

// cmdUsage displays the usage of the app
func cmdUsage() {
	fmt.Print("Usage: yapi [OPTION]...\n\n")
	fmt.Printf("yapi - Yet Another Pipe Implementation - v%s\n", YAPI_VERSION)

	fmt.Printf("\nOptions:\n")
	//flag.PrintDefaults()
	flag.VisitAll(func(flg *flag.Flag) {
		defVal := ""
		if flg.DefValue != "" && flg.DefValue != "false" {
			defVal = "(default: " + flg.DefValue + ")"
		}
		fmt.Printf("  -%-10s : %s %s\n", flg.Name, flg.Usage, defVal)
	})

	fmt.Print("\nExamples:\n")
	fmt.Print("  yapi -cc ls\n")
	fmt.Print("  yapi -pc /path/pipe.json -cc \"tail -f /var/log/syslog\"\n")
	fmt.Print("  yapi -cc \"top -b -n 1\" | grep ssh\n")

	fmt.Printf("\nPlease report issues to https://github.com/cmfatih/yapi/issues\n")
}

// cmdVer displays the version information of the app
func cmdVer() {
	fmt.Printf("yapi version %s\n", YAPI_VERSION)
}

// flagCN parses `cn` flag.
func flagCN(cn string) []string {
	if cn != "" {
		var cns []string
		spl := strings.Split(cn, ",")
		for _, val := range spl {
			val = strings.TrimSpace(val)
			if val != "" {
				cns = append(cns, val)
			}
		}

		return cns
	}

	return nil
}
