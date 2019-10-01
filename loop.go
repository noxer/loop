package loop

import (
	"runtime"
	"sync"
)

func init() {
	runtime.LockOSThread()
}

var (
	m  sync.Mutex
	ch chan func()
	r  bool
)

// MainBuffered starts the main loop with a buffered queue. This should be
// called from the main() function. It will block until Terminate is called.
func MainBuffered(main func(), buffer uint) {
	m.Lock()
	if ch != nil {
		panic("main loop is already running")
	}

	ch = make(chan func(), buffer)
	r = true
	go main()

	ch <- m.Unlock
	for f := range ch {
		f()
	}

	m.Lock()
	ch = nil
	m.Unlock()
}

// Main creates a main loop with an unbuffered queue. This should be called
// from the main() function. Attempting to start a second main loop results in
// a panic.
func Main(main func()) {
	MainBuffered(main, 0)
}

// Terminate terminates a running main loop. If no main loop is running this is a no-op.
func Terminate() {
	m.Lock()
	if !r {
		// Already terminated
		m.Unlock()
		return
	}

	r = false
	close(ch)
	m.Unlock()
}

// Schedule sends a function to the main loop to be executed. The function
// returns false iff the main loop is not running.
func Schedule(f func()) bool {
	m.Lock()
	if !r {
		m.Unlock()
		return false
	}

	ch <- f
	m.Unlock()

	return true
}

// ScheduleAwait send a function to the main loop and waits until it has been
// executed. The function returns false iff the main loop is not running.
func ScheduleAwait(f func()) bool {
	c := make(chan struct{})
	r := Schedule(func() {
		f()
		close(c)
	})
	if !r {
		return false
	}

	<-c
	return true
}

// IsRunning indicates if the main looper is currently running.
// This is just a hint that this package is in use. The main loop may have been
// started or terminated once you attempt your action.
func IsRunning() bool {
	m.Lock()
	running := r
	m.Unlock()

	return running
}
