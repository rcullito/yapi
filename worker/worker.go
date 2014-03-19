// yapi
// Copyright (c) 2014 Fatih Cetinkaya (http://github.com/cmfatih/yapi)
// For the full copyright and license information, please view the LICENSE.txt file.

// Package worker provides concurrency (goroutines, channels, etc.) related stuff.
package worker

import (
	"code.google.com/p/go-uuid/uuid"
	"errors"
)

var (
	workers     = map[string]Worker{}
	workerKinds = map[string]bool{"cce": true}
)

// Worker is the interface that must be implemented by workers.
type Worker interface {

	// ID returns the unique id of the worker.
	ID() string

	// Kind returns the kind of the worker.
	Kind() string

	// SetOptions sets the options of the worker.
	SetOptions(workerOpts WorkerOptions) error

	// Start starts the worker.
	Start() error
}

// WorkerOptions implements the options of the worker.
// Putty represents worker's distinctive options.
type WorkerOptions struct {
	Putty interface{}
}

// New returns a new worker with the given kind.
func New(workerKind string) (Worker, error) {

	// Check vars
	if workerKind == "" || workerKinds[workerKind] != true {
		return nil, errors.New("invalid kind (" + workerKind + ")")
	}

	// Init worker
	workerID := uuid.New()

	if workerKind == "cce" {
		worker := cceWorker{
			id:   workerID,
			kind: workerKind,
		}

		// Add to the list
		workers[workerID] = &worker

		return &worker, nil
	}

	return nil, errors.New("unexpected error! (worker.New)")
}

// Start starts the worker by the given worker id.
func Start(workerID string) error {

	// Get the worker
	if workerID == "" || workers[workerID] == nil {
		return errors.New("invalid worker id (" + workerID + ")")
	}

	// Start the worker
	return workers[workerID].Start()
}
