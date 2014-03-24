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
	"time"
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

	// Init timeout vars
	var timeout <-chan time.Time
	if wCCE.options.Timeout > 0 {
		timeout = time.After(time.Duration(wCCE.options.Timeout) * time.Millisecond)
	}

	// Init sync
	wg := new(sync.WaitGroup)
	cliCnt := len(wCCE.options.Clients)
	wg.Add(cliCnt)

	if wCCE.options.Method == "serial" {

		// Init channel
		channDone := make(chan bool)

		go func() {
			for _, name := range wCCE.options.Clients {
				if err := client.ExecCmd(wCCE.options.Cmd, name); err != nil {
					if wCCE.options.CmdErrPrint == true {
						fmt.Println("failed to execute the command: " + err.Error())
					}
				}
				wg.Done()
			}
			channDone <- true
		}()

		if wCCE.options.Timeout > 0 {
			select {
			case _ = <-channDone:
				return nil
			case <-timeout:
				if wCCE.options.CmdErrPrint == true {
					fmt.Println("failed to execute the command: timeout (" + fmt.Sprintf("%d", wCCE.options.Timeout) + "ms)")
				}
				return nil
			}
		}
	} else if wCCE.options.Method == "parallel" {

		// Init channel
		channDone := make(chan int)

		go func() {
			for i := 0; i < cliCnt; i++ {
				go func(cliName string, index int) {
					err := client.ExecCmd(wCCE.options.Cmd, cliName)
					if err != nil {
						if wCCE.options.CmdErrPrint == true {
							fmt.Println("failed to execute the command: " + err.Error())
						}
					}
					wg.Done()
					channDone <- index + 1
				}(wCCE.options.Clients[i], i)
			}
		}()

		if wCCE.options.Timeout > 0 {
			for {
				select {
				case val := <-channDone:
					if val == cliCnt {
						return nil
					}
				case <-timeout:
					if wCCE.options.CmdErrPrint == true {
						fmt.Println("failed to execute the command: timeout (" + fmt.Sprintf("%d", wCCE.options.Timeout) + "ms)")
					}
					return nil
				}
			}
		}
	} else {
		if wCCE.options.CmdErrPrint == true {
			fmt.Println("invalid client command execution method: " + wCCE.options.Method)
			return nil
		}
	}

	wg.Wait()

	return nil
}

// CCEOptions implements the CCE options.
type CCEOptions struct {
	Clients     []string
	Cmd         string
	CmdErrPrint bool
	Method      string
	Timeout     int64
}
