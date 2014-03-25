// yapi
// Copyright (c) 2014 Fatih Cetinkaya (http://github.com/cmfatih/yapi)
// For the full copyright and license information, please view the LICENSE.txt file.

// yapi - Yet Another Pipe Implementation
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/cmfatih/yapi/client"
	"github.com/cmfatih/yapi/pipe"
	"github.com/cmfatih/yapi/worker"
	"os"
	"os/user"
	"runtime"
	"runtime/pprof"
	"strings"
)

const (
	YAPI_VERSION = "0.3.3^HEAD" // app version
)

var (
	gvCLEC      string    // command line escape char
	gvHOME      string    // User home directory
	gvPipeConf  pipe.Conf // pipe config
	gvCliNames  []string  // client names
	gvCliGroups []string  // client groups

	flPipeConf string // pipe config flag
	flCliName  string // client name flag
	flCliGroup string // client group flag
	flCliCmd   string // client command flag
	flCliCEM   string // client command execution method flag
	flCliCET   int64  // client command execution timeout
	flSSH      string // simple ssh client flag
	flHelp     bool   // help flag
	flVersion  bool   // version flag
	flDbg      bool   // debug flag
	flProfCPU  string // cpu profile flag
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
	flag.StringVar(&flPipeConf, "pc", "", "Pipe configuration file. Default; pipe.json")

	flag.StringVar(&flCliCmd, "cc", "", "Client command that will be executed.")
	flag.StringVar(&flCliName, "cn", "", "Client name(s) those will be connected.")
	flag.StringVar(&flCliGroup, "cg", "", "Client group name(s) those will be connected.")
	flag.StringVar(&flCliCEM, "ccem", "serial", "Execution method for client command. Default; serial")
	flag.Int64Var(&flCliCET, "ccet", 0, "Timeout (millisecond) for client command execution.")

	flag.StringVar(&flSSH, "ssh", "", "Simple SSH client command execution.")

	flag.BoolVar(&flHelp, "help", false, "Display help and exit.")
	flag.BoolVar(&flHelp, "h", false, "Display help and exit.")
	flag.BoolVar(&flVersion, "version", false, "Display version information and exit.")
	flag.BoolVar(&flVersion, "v", false, "Display version information and exit.")
	flag.BoolVar(&flDbg, "dbg", false, "Display debug information end exit.")
	flag.StringVar(&flProfCPU, "profcpu", "", "Write cpu profile to file.")
}

func main() {

	// Init flags
	flag.Parse()

	// Profile cpu
	if flProfCPU != "" {
		f, err := os.Create(flProfCPU)
		if err != nil {
			fmt.Printf("Failed to profile cpu: %s\n", err)
			return
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Printf("Failed to profile cpu: %s\n", err)
			return
		}
		defer pprof.StopCPUProfile()
		//defer f.Close() // do not defer
	}

	// Version
	if flVersion == true {
		flagVer()
		return
	}

	// Help
	if flHelp == true {
		flagHelp()
		return
	}

	// Debug
	if flDbg == true {
		flagDbg()
		return
	}

	// Simple SSH CCE
	if flSSH != "" {
		if err := flagSSH(flSSH, flCliCmd, flCliCEM, flCliCET); err != nil {
			fmt.Println(err.Error())
		}
		return
	}

	// Client command
	if flCliCmd != "" {
		// pipe config
		if err := flagPC(flPipeConf); err != nil {
			fmt.Println(err.Error())
			return
		}

		// client names and groups
		flagCNG(flCliName, flCliGroup)

		// client command
		if err := flagCC(flCliCmd, flCliCEM, flCliCET, gvCliNames); err != nil {
			fmt.Println(err.Error())
			return
		}

		return
	}

	// Default
	flagHelp()
	return
}

// flagVer displays the version information.
func flagVer() {

	// Output
	fmt.Print("Version:\n\n")
	fmt.Printf("  yapi : %s\n", YAPI_VERSION)
}

// flagHelp displays the help.
func flagHelp() {

	// Output
	fmt.Print("Usage: yapi [OPTION]...\n\n")
	fmt.Printf("yapi - Yet Another Pipe Implementation - v%s\n", YAPI_VERSION)
	fmt.Print(`
  Options:
    -pc           : Pipe configuration file. Default; pipe.json

    -cc           : Client command that will be executed.
    -cn           : Client name(s) those will be connected.
                    Use comma (,) for multi-client.
    -cg           : Client group name(s) those will be connected.
                    Use comma (,) for multi-group.
    -ccem         : Execution method for client command. Default; serial
                    Possible values; serial (~), parallel (//)
    -ccet         : Timeout (millisecond) for client command execution.

    -ssh          : Simple SSH client command execution.
                    It uses the current/given username and HOME/.ssh/id_rsa
                    for the private key file.
                    Syntax: [user@]host[:22]

    -h, -help     : Display help and exit.
    -v, -version  : Display version information and exit.
    -dbg          : Display debug information end exit.
    -profcpu      : Write cpu profile to file.

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
}

// flagDbg displays the version information.
func flagDbg() {

	// Init vars
	username := ""
	if u, err := user.Current(); err != nil {
		username = err.Error()
	} else {
		username = u.Username
	}

	// Output
	fmt.Print("Debug:\n\n")
	fmt.Printf("  yapi version : %s\n", YAPI_VERSION)
	fmt.Printf("  platform     : %s\n", runtime.GOOS)
	fmt.Printf("  username     : %s\n", username)
	fmt.Printf("  home         : %s\n", gvHOME)
}

// flagPC loads pipe config.
func flagPC(pcFile string) error {

	if err := gvPipeConf.Load(pcFile, pipe.LoadOpt{CliInit: true}); err != nil {
		return errors.New("Error due pipe configuration: " + err.Error())
	}

	return nil
}

// flagCNG sets the client names and groups.
func flagCNG(cliName, cliGroup string) {

	// Init vars
	gvCliNames = flagMultiParser(cliName, ",")
	gvCliGroups = flagMultiParser(cliGroup, ",")

	// Check groups
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

// flagCC executes the client command.
func flagCC(cliCmd, cliCmdEM string, cliCmdET int64, cliNames []string) error {

	// Default client
	if cliNames == nil {
		if _, name := gvPipeConf.CliDef(); name != "" {
			cliNames = append(cliNames, name)
		}
	}

	// Create a new worker
	ccew, err := worker.New("cce")
	if err != nil {
		return errors.New("Failed to execute the command: " + err.Error())
	}

	// Set the options
	if err := ccew.SetOptions(
		worker.WorkerOptions{
			Putty: worker.CCEOptions{
				Clients:     cliNames,
				Cmd:         cliCmd,
				CmdErrPrint: true,
				Method:      flagSymbolParser(cliCmdEM),
				Timeout:     cliCmdET,
			},
		},
	); err != nil {
		return errors.New("Failed to execute the command: " + err.Error())
	}

	// Start the worker
	if err := ccew.Start(); err != nil {
		return errors.New("Failed to execute the command: " + err.Error())
	}

	return nil
}

// flagSSH executes the given command via ssh client.
func flagSSH(sshOpt, cliCmd, cliCmdEM string, cliCmdET int64) error {

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
			return errors.New("Unexpected error: " + err.Error())
		} else {
			cliConfs = append(cliConfs, string(buf))
			cliNames = append(cliNames, name)
		}
	}

	jsonCont := "{\"Clients\":[" + strings.Join(cliConfs, ",") + "]}"

	if err := gvPipeConf.LoadJSON(jsonCont, pipe.LoadOpt{CliInit: true}); err != nil {
		return errors.New("Error due pipe configuration: " + err.Error())
	}

	if err := flagCC(cliCmd, cliCmdEM, cliCmdET, cliNames); err != nil {
		return err
	}

	return nil
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
