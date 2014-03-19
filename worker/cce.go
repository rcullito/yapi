// yapi
// Copyright (c) 2014 Fatih Cetinkaya (http://github.com/cmfatih/yapi)
// For the full copyright and license information, please view the LICENSE.txt file.

// This file contains implementation for client command execution (CCE) worker.

package worker

import (
	"errors"
	"fmt"
	"github.com/cmfatih/yapi/client"
)

var (
	cceMethods = map[string]bool{"serial": true, "parallel": true}
)

// cceWorker implements a CCE worker.
type cceWorker struct {
	id      string     // id
	kind    string     // kind of worker (cce)
	options CCEOptions // options
}

// ID returns the unique id of the worker.
func (wCCE *cceWorker) ID() string {
	return wCCE.id
}

// Kind returns the kind of the worker.
func (wCCE *cceWorker) Kind() string {
	return wCCE.kind
}

// SetOptions sets the options of the worker.
func (wCCE *cceWorker) SetOptions(workerOpts WorkerOptions) error {

	// Check the options
	var cceOpts CCEOptions
	cceOpts = workerOpts.Putty.(CCEOptions)

	if cceOpts.Clients == nil {
		return errors.New("clients option is missing")
	} else if cceOpts.Cmd == "" {
		return errors.New("command is missing")
	} else if cceOpts.Method == "" || cceMethods[cceOpts.Method] != true {
		return errors.New("invalid method option (" + cceOpts.Method + ")")
	} else if cceOpts.Method == "parallel" {
		return errors.New("parallel method is still under development...")
	}

	wCCE.options = cceOpts

	return nil
}

// Start starts the worker.
func (wCCE *cceWorker) Start() error {

	// Check the options
	if len(wCCE.options.Clients) == 0 {
		return errors.New("there is no client to work on")
	}

	// Init channels
	cliCnt := len(wCCE.options.Clients)
	cliList := make(chan string, cliCnt)
	jobDone := make(chan bool)

	go func() {
		for {
			name, more := <-cliList
			if more {
				if err := client.ExecCmd(wCCE.options.Cmd, name); err != nil {
					if wCCE.options.CmdErrPrint == true {
						fmt.Println("Failed to execute the command: " + err.Error())
					}
				}
			} else {
				jobDone <- true
				return
			}
		}
	}()

	for _, name := range wCCE.options.Clients {
		cliList <- name
	}
	close(cliList)

	<-jobDone

	return nil
}

// CCEOptions implements the CCE options.
type CCEOptions struct {
	Clients     []string
	Cmd         string
	CmdErrPrint bool
	Method      string
}
