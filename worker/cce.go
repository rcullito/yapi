// yapi
// Copyright (c) 2014 Fatih Cetinkaya (http://github.com/cmfatih/yapi)
// For the full copyright and license information, please view the LICENSE.txt file.

// This file contains implementation for client command execution (CCE) worker.

package worker

import (
	"errors"
	"fmt"
	"github.com/cmfatih/yapi/client"
	"sync"
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
		return errors.New("there is no any client to use")
	} else if cceOpts.Cmd == "" {
		return errors.New("missing client command")
	} else if cceOpts.Method == "" || cceMethods[cceOpts.Method] != true {
		return errors.New("invalid client command execution method (" + cceOpts.Method + ")")
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

	// Because of the ssh pkg; this code block is the best approach so far.
	// A channel, a wait group and goroutines.
	// Be careful and test it well during implement timeout stuff.

	// Init sync
	chann := make(chan bool)
	wg := new(sync.WaitGroup)
	cliCnt := len(wCCE.options.Clients)
	wg.Add(cliCnt)

	if wCCE.options.Method == "serial" {

		// TODO: Implement timeout

		go func() {
			for _, name := range wCCE.options.Clients {
				if err := client.ExecCmd(wCCE.options.Cmd, name); err != nil {
					if wCCE.options.CmdErrPrint == true {
						fmt.Println("failed to execute the command: " + err.Error())
					}
				}
				wg.Done()
			}
			chann <- true
		}()

	} else if wCCE.options.Method == "parallel" {

		// TODO: Implement timeout

		go func() {
			for i := 0; i < cliCnt; i++ {
				go func(cliName string) {
					err := client.ExecCmd(wCCE.options.Cmd, cliName)
					if err != nil {
						if wCCE.options.CmdErrPrint == true {
							fmt.Println("failed to execute the command: " + err.Error())
						}
					}

					wg.Done()
				}(wCCE.options.Clients[i])
			}
			chann <- true
		}()
	} else {
		if wCCE.options.CmdErrPrint == true {
			fmt.Println("invalid client command execution method: " + wCCE.options.Method)
		}
	}

	wg.Wait()

	<-chann

	return nil
}

// CCEOptions implements the CCE options.
type CCEOptions struct {
	Clients     []string
	Cmd         string
	CmdErrPrint bool
	Method      string
}
