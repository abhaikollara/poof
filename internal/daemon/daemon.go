package daemon

import (
	"fmt"
	"os"
	"time"

	"abhai.dev/poof/internal/registry"
	"abhai.dev/poof/internal/sweep"
)

const PollInterval = 10 * time.Second

// Run is the daemon main loop. It blocks until the stop channel is closed.
func Run(stop <-chan struct{}) {
	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()

	doSweep()

	for {
		select {
		case <-ticker.C:
			doSweep()
		case <-stop:
			return
		}
	}
}

func doSweep() {
	lock, err := registry.Lock()
	if err != nil {
		fmt.Fprintf(os.Stderr, "poof-daemon: lock: %v\n", err)
		return
	}
	defer registry.Unlock(lock)

	reg, err := registry.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "poof-daemon: load: %v\n", err)
		return
	}

	before := len(reg.Entries)
	sweep.Run(reg, time.Now().UTC())
	after := len(reg.Entries)

	if before != after {
		if err := registry.Save(reg); err != nil {
			fmt.Fprintf(os.Stderr, "poof-daemon: save: %v\n", err)
		}
	}
}
