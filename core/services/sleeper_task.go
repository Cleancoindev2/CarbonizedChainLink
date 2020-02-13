package services

import (
	"sync"
)

// SleeperTask represents a task that waits in the background to process some work.
type SleeperTask interface {
	Start() error
	Stop() error
	WakeUp()
}

// Worker is a simple interface that represents some work to do repeatedly
type Worker interface {
	Work()
}

type sleeperTask struct {
	worker   Worker
	waker    chan struct{}
	closer   chan struct{}
	workerWG sync.WaitGroup
}

// NewSleeperTask takes a worker and returns a SleeperTask.
//
// SleeperTask is guaranteed to call Work on the worker at least once for every
// WakeUp call.
// If the Worker is busy when WakeUp is called, the Worker will be called again
// immediately after it is finished. For this reason you should take care to
// make sure that Worker is idempotent.
// WakeUp does not block.
//
func NewSleeperTask(worker Worker) SleeperTask {
	return &sleeperTask{
		worker: worker,
		waker:  make(chan struct{}, 1),
		closer: make(chan struct{}, 1),
	}
}

// Start begins the SleeperTask
func (s *sleeperTask) Start() error {
	go s.workerLoop()
	return nil
}

// Stop stops the SleeperTask
func (s *sleeperTask) Stop() error {
	s.closer <- struct{}{}
	s.workerWG.Wait()
	return nil
}

// WakeUp wakes up the sleeper task, asking it to execute its Worker.
func (s *sleeperTask) WakeUp() {
	s.workerWG.Add(1)
	select {
	case s.waker <- struct{}{}:
	default:
		s.workerWG.Done()
	}
}

// workerLoop is the goroutine behind the sleeper task that waits for a signal
// before kicking off the worker
func (s *sleeperTask) workerLoop() {
	for {
		select {
		case <-s.waker:
			s.worker.Work()
			s.workerWG.Done()
		case <-s.closer:
			return
		}
	}
}
